@description('Optional. The app registration used to authenticate.')
param appRegistration string = '18af2af2-1b65-405d-a31c-c49fd740bcaa'

@description('Optional. The port the container is using, default = 8080.')
param containerPort int = 8080

@description('Required. The image to run.')
param image string

@description('Optional. Location to deploy resources in, override if necessary, use default in most cases')
param location string = resourceGroup().location

@minLength(3)
@maxLength(12)
@description('Optional. Name of the app.')
param name string = 'guard-portal'

@description('Required. The new revision suffix.')
param revisionSuffix string

@description('Required. The new/updated traffic configuration.')
param traffic array

var uniqueId = uniqueString(subscription().id, resourceGroup().id, location)
var uniqueProjectName = '${name}-${uniqueId}'
var uniqueProjectNameAlphanumeric = replace(uniqueProjectName, '-', '')

var tags = resourceGroup().tags

resource cr 'Microsoft.ContainerRegistry/registries@2023-01-01-preview' existing = {
  name: uniqueProjectNameAlphanumeric
}

resource caEnv 'Microsoft.App/managedEnvironments@2023-05-01' existing = {
  name: 'caenv-${uniqueProjectName}'
}

resource mi 'Microsoft.ManagedIdentity/userAssignedIdentities@2023-01-31' existing = {
  name: 'mi-${uniqueProjectName}'
}

resource kv 'Microsoft.KeyVault/vaults@2023-02-01' existing = {
  name: take('kv-${uniqueProjectName}', 24)
}

resource containerApp 'Microsoft.App/containerApps@2023-05-01' = {
  name: name
  location: location
  tags: tags
  identity: {
    type: 'UserAssigned'
    userAssignedIdentities: {
      '${mi.id}': {}
    }
  }
  properties: {
    configuration: {
      activeRevisionsMode: 'Multiple'
      dapr: {
        enabled: false
      }
      ingress: {
        allowInsecure: false
        external: true
        targetPort: containerPort
        traffic: traffic
        transport: 'auto'
      }
      maxInactiveRevisions: 10
      registries: [
        {
          identity: mi.id
          server: cr.properties.loginServer
        }
      ]
      secrets: [
        {
          name: 'microsoft-provider-authentication-secret'
          keyVaultUrl: '${kv.properties.vaultUri}secrets/microsoft-provider-authentication-secret'
          identity: mi.id
        }
      ]
    }
    environmentId: caEnv.id
    template: {
      containers: [
        {
          env: [
            {
              name: 'AUTH_HEADER'
              value: 'X-Ms-Client-Principal'
            }
            {
              name: 'FILES_BASE_PATH'
              value: '/guard-portal'
            }
            {
              name: 'FILES_GENERAL_DIR'
              value: 'algemeen'
            }
             {
              name: 'FILES_SAFETY_DIR'
              value: 'safety'
            }
            {
              name: 'FILES_SCHEDULES_DIR'
              value: 'planningen'
            }
            {
              name: 'SERVER_ADDRESS'
              value: ':${containerPort}'
            }
            {
              name: 'SERVER_CSP'
              value: 'default-src \'self\'; script-src \'self\' \'unsafe-inline\'; style-src \'self\' \'unsafe-inline\'; frame-ancestors \'none\';'
            }
            // {
            //   name: 'SERVER_HSTS_MAX_AGE'
            //   value: '31536000'
            // }
            // {
            //   name: 'SERVER_HSTS_INCLUDE_SUBDOMAINS'
            //   value: 'true'
            // }
            // {
            //   name: 'SERVER_HSTS_PRELOAD'
            //   value: 'true'
            // }
          ]
          image: image
          name: name
          resources: {
            cpu: json('0.25')
            memory: '0.5Gi'
          }
          volumeMounts: [
            {
              mountPath: '/guard-portal'
              volumeName: 'fileshare-guard-portal'
            }
          ]
        }
      ]
      revisionSuffix: revisionSuffix
      scale: {
        minReplicas: 1
        maxReplicas: 2
      }
      volumes: [
        {
          name: 'fileshare-guard-portal'
          storageName: 'fileshare-guard-portal'
          storageType: 'AzureFile'
        }
      ]
    }    
  }  
}

resource authConfig 'Microsoft.App/containerApps/authConfigs@2023-05-01' = {
  name: 'current'
  parent: containerApp
  properties: {
    globalValidation: {
      redirectToProvider: 'azureActiveDirectory'
      unauthenticatedClientAction: 'RedirectToLoginPage'
    }
    identityProviders: {
      azureActiveDirectory: {
        enabled: true
        login: {
          disableWWWAuthenticate: true
        }
        registration: {
          clientId: appRegistration
          clientSecretSettingName: 'microsoft-provider-authentication-secret'
          openIdIssuer: '${environment().authentication.loginEndpoint}${subscription().tenantId}/v2.0'
        }
      }
    }
    platform: {
      enabled: true
    }
  }
}
