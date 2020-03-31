#!/usr/bin/env python2.7

from __future__ import print_function
from pprint import pprint
import sys
import os
import re
import itertools as it
import tempfile
import argparse
from collections import OrderedDict

""" Add support for ordered yaml loading in pythong yaml
"""
import yaml as pyyaml

_items = 'viewitems' if sys.version_info < (3,) else 'items'


def map_representer(dumper, data):
    return dumper.represent_dict(getattr(data, _items)())


def map_constructor(loader, node):
    loader.flatten_mapping(node)
    return OrderedDict(loader.construct_pairs(node))


if pyyaml.safe_dump is pyyaml.dump:
    # PyYAML v4.1
    SafeDumper = pyyaml.dumper.Dumper
    DangerDumper = pyyaml.dumper.DangerDumper
    SafeLoader = pyyaml.loader.Loader
    DangerLoader = pyyaml.loader.DangerLoader
else:
    SafeDumper = pyyaml.dumper.SafeDumper
    DangerDumper = pyyaml.dumper.Dumper
    SafeLoader = pyyaml.loader.SafeLoader
    DangerLoader = pyyaml.loader.Loader

pyyaml.add_representer(dict, map_representer, Dumper=SafeDumper)
pyyaml.add_representer(OrderedDict, map_representer, Dumper=SafeDumper)
pyyaml.add_representer(dict, map_representer, Dumper=DangerDumper)
pyyaml.add_representer(OrderedDict, map_representer, Dumper=DangerDumper)


if sys.version_info < (3, 7):
    pyyaml.add_constructor('tag:yaml.org,2002:map', map_constructor, Loader=SafeLoader)
    pyyaml.add_constructor('tag:yaml.org,2002:map', map_constructor, Loader=DangerLoader)


del map_constructor, map_representer


# Merge PyYAML namespace into ours.
# This allows users a drop-in replacement:
#   import oyaml as yaml
# from yaml import *
import yaml


regex_attr = "\\\.([a-zA-Z0-9\_]+)"
""" str: matches attribute names in values, e.g, \.region, \.gcp_region, etc.
Attribute names must include letters or underscores only.
"""

regex_env = "\$\{([a-zA-Z0-9_]+)\}"
""" str: matches environment variable names, e.g, ${REGION}, ${GCP_REGION}, etc.
Environment variables must be include letters, numbers, and underscores only.
"""

def panic(msg):
  print(msg, file=sys.stderr)
  sys.exit(1)


def resolve(match, *objects):
  """Given an array of objects and a regex match, this function returns the first
  matched group if it exists in one of the objects, otherwise returns the orginial
  fully matches string by the regex.

  Example: if regex = \\\.([a-z]) and string = test\.abc, then
  the match = {group0: \.abc, group1: abc}. Assuimg one object:
  - obj = {abc: def}, then we return 'def'
  - obj = {test: value}, then we return \.abc

  Args:
      objects (array[dict]): the array of objects we use to look up the key in match.group(1)
      match: the regex match object

  Returns:
      str: the value of the matched group(1) in the first object found if exists, otherwise
           returns the fully matched string.
  """
  for obj in objects:
    if obj is not None and match.group(1) in obj:
      return str(obj[match.group(1)])
  return match.group(0)


def resolve_placeholders(regex, item, resolvers):
  """This function resolves placeholders in a record with its values. Example:
  if item = {"aws_region": "us-east-1", "region": "\.aws_region"} and resolvers is [item], then
  the output is {"aws_region": "us-east-1", "region": "us-east-1"}.

  Args:
      regex (str): the regex to use for substitution for values in the record
      item (dict): the object to resolve its values
      resolvers (array[dict]): an array of records to use for resolution

  """
  _resolve = lambda match: resolve(match, *resolvers)

  for key in item.keys():
    cond = True
    while cond:
      value = item[key]
      item[key] = re.sub(regex, _resolve, str(value))
      cond = (item[key] != value)


def add_prefix(obj, prefix):
  """This function adds prefix to keys in obj

  Args:
      obj (dict): object with keys/values
      prefix (str): string value to prefix

  Returns:
      dict: new object with prefix addeds
  """
  res = OrderedDict()
  for key, val in obj.iteritems():
    res[concat_parts(prefix, key)] = val

  return res


def concat_parts(*args):
  """Concat a list of strings using '_' ignoring None values

  Args:
      args (array[str]): a list of strings

  Returns:
      str: the concatenated string result
  """
  name = []
  for a in args:
    if a is not None:
      name.append(a.lower().replace('-', '_'))
  return '_'.join(name)


def get_override(section):
  override = OrderedDict()
  override.update(section)
  if 'metadata' in override:
    del override['metadata']
  if 'default' in override:
    del override['default']
  return override


def load_section(section, overrides):
  """Load the given section and resolve using the give list of overrides

  Args:
      section (dict): a section from the yaml file
      overrides (list[str]): a list of overrides, e.g., [gcp, prod, us-central1, cluster1]

  Returns:
      dict: a dict with two keys 'metadata', and 'default' for the resolves section metadata and default values.
  """
  metadata = section['metadata'] if 'metadata' in section else OrderedDict()
  default = section['default'] if 'default' in section else OrderedDict()
  override = get_override(section)

  # Loop over all input overrides, e.g., [gcp, prod, us-central1]
  for o in overrides:
    # If we don't have an override section for the current object, we stop here
    if o not in override:
      break

    # update the override object, e.g., override['gcp']
    override = override[o]

    # override metadata values if exists
    metadata_override = override['metadata'] if 'metadata' in override else OrderedDict()
    metadata.update(metadata_override)

    # override default values if exists
    default_override = override['default'] if 'default' in override else OrderedDict()
    default.update(default_override)

    # move to the next override if exists
    override = get_override(override)


  for key in metadata.keys():
    if metadata[key] is None:
      del metadata[key]

  for key in default.keys():
    if default[key] is None:
      del default[key]

  return {'metadata': metadata, 'default': default}


def load(ifname, overrides):
  with open(ifname, 'r') as fin:
    yaml_obj = yaml.load(fin, Loader=SafeLoader)

  sections = []

  # Load section and resolve overrides
  for section_name, section_info in yaml_obj.iteritems():
    section = load_section(section_info, overrides)
    sections.append(section)

  # Resolve placeholders in values
  for i in xrange(len(sections)):  # Loop over all sections
    section = sections[i]['default']
    for j in xrange(i, -1, -1):  # Loop over sections before starting by myself
      resolver = sections[j]['default']
      # replace the placeholders in the values for this object
      resolve_placeholders(regex_attr, section, [resolver])

  for section in sections:
    # preprend the prefix to all keys if exists in metadata
    section['default'] = add_prefix(section['default'], section['metadata'].get('prefix'))

  # Convert keys to upper case
  for section in sections:
    new_section = OrderedDict()
    for key, val in section['default'].iteritems():
      new_section[key.upper()] = val
    section['default'] = new_section

  # Flatten the array into an object
  res = OrderedDict()
  for section in sections:
    for key, val in section['default'].iteritems():
      res[key] = val

  # Override using environment variables before resolving environment
  # variables in the values
  for key in res.keys():
    if key in os.environ:
      res[key] = os.environ[key]


  # Resolve environment variables
  resolve_placeholders(regex_env, res, [os.environ, res])

  return res


def dump_as_exports(res):
  """Dump the result object to stdout

  Args:
      res (dict): the object to dump to the output file
  """
  for k in res.iterkeys():
    print("export {}={}".format(k, res[k]), file=sys.stdout)


def main(args):
  parser = argparse.ArgumentParser(description="Config parser for yaml properties")
  parser.add_argument('--files', '-f', nargs='+', required=True, type=str)
  parser.add_argument('--overrides', '-o', nargs='*', type=str)

  args = parser.parse_args(args)

  config_files = args.files
  overrides = args.overrides if args.overrides else []

  # load the intput files for the given overrides.
  configs = map(lambda f: load(f, overrides), config_files)

  res = OrderedDict()

  for config in configs:
    for key, val in config.iteritems():
      if key not in res:
        res[key] = val

  # dump the exports to output file, and print out the output file name.
  dump_as_exports(res)


if __name__ == "__main__":
  main(sys.argv[1:])
