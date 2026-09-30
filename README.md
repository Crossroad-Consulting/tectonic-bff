# WebApp

## Create container app (or update it manually)

1. Log into Azure
```sh
# Login to your Azure Active Directory tenant
az login

# Make sure you are using the right subscription
az account show
az account list

# If you are not in the correct subscription, change it substituting SUBSCRIPTIONID with the proper subscription id
az account set --subscription {SUBSCRIPTIONID}
```

2. Deploy bicep
```sh
export group="<name resource-group>"
export cr="<name container registry>.azurecr.io"
export image="guard-portal/webapp"

az deployment group create --resource-group $group --template-file deploy/container-app.bicep \
  --parameters \
    image='$cr/$image:1.0.0' \
    revisionSuffix='' \
    traffic='[{"latestRevision": true,"weight": 100}]' \
  --confirm-with-what-if
```