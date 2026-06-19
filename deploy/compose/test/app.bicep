extension radius

@description('The Radius application ID. Injected automatically by the rad CLI.')
param application string

@description('The Radius environment ID. Injected automatically by the rad CLI.')
param environment string

// A sample "to-do list" web app container that connects to the Redis cache below.
// The connection injects the Redis hostname/port/password into the container as
// environment variables, so the app discovers its database with no hard-coded values.
resource todo 'Applications.Core/containers@2023-10-01-preview' = {
  name: 'todo'
  properties: {
    application: application
    container: {
      image: 'ghcr.io/radius-project/samples/demo:latest'
      ports: {
        web: {
          containerPort: 3000
        }
      }
    }
    connections: {
      redis: {
        source: db.id
      }
    }
  }
}

// A Redis cache provisioned by the environment's recipe. The recipe deploys a
// Redis pod and service into the k3s cluster through the Deployment Engine.
resource db 'Applications.Datastores/redisCaches@2023-10-01-preview' = {
  name: 'db'
  properties: {
    application: application
    environment: environment
  }
}
