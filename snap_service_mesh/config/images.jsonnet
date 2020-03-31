{
  registry: {
    gcp: {
      publishTo: [
        "gcr.io/api-gateway-dev",
        "gcr.io/snap-mesh-images",
        "us.gcr.io/snap-mesh-images",
        "eu.gcr.io/snap-mesh-images",
        "asia.gcr.io/snap-mesh-images",
      ],
      pullFrom: {
        // Default region to pull from for Jenkins builds
        "default": "gcr.io/api-gateway-dev",
        // Regions to pull from for Gateway deployments
        "us-central1": "us.gcr.io/snap-mesh-images",
        "us-east1": "us.gcr.io/snap-mesh-images",
        "us-east4": "us.gcr.io/snap-mesh-images",
        "europe-west1": "eu.gcr.io/snap-mesh-images",
        "asia-southeast1": "asia.gcr.io/snap-mesh-images",
      },
    },
    aws: {
      publishTo: [
        "729677352544.dkr.ecr.us-west-2.amazonaws.com",
        "729677352544.dkr.ecr.us-east-1.amazonaws.com",
        "729677352544.dkr.ecr.eu-west-1.amazonaws.com",
        "729677352544.dkr.ecr.ap-southeast-1.amazonaws.com",
      ],
      pullFrom: {
        // Default region to pull from for Jenkins builds
        "default": "729677352544.dkr.ecr.us-east-1.amazonaws.com",
        // Regions to pull from for Gateway deployments
        "us-west-2": "729677352544.dkr.ecr.us-west-2.amazonaws.com",
        "us-east-1": "729677352544.dkr.ecr.us-east-1.amazonaws.com",
        "eu-west-1": "729677352544.dkr.ecr.eu-west-1.amazonaws.com",
        "ap-southeast-1": "729677352544.dkr.ecr.ap-southeast-1.amazonaws.com",
      },
    },
  },
}
