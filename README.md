# HELM Key Protect Postrenderer

With this helm v3 post-renderer plugin, you can have encrypted strings in your chart and they will be decrypted and made part of the manifest before it gets deployed to kubernetes.


## Example

Starting out from a helm chart top-level directory, let's say you have a _values.yaml_ file like this one:
```yaml
auth_token: eyJjaXBoZXJ0ZXh0IjoibHAxdXNzOGo1UW1qTUt2SkVBTmlVUmVVczMwPSIsIml2IjoiRVlWcStsUjlvNGZDVlJpNyIsInZlcnNpb24iOiI0LjAuMCIsImhhbmRsZSI6ImJhNGQ0NDM3LTFkY2EtNGI3Yi1iY2ExLTNlZmZhYjg5ODc5OCJ9
```

`auth_token` is a secret and you should not store these kinds of values with a deployment chart. However, in this case it has already been securely encrypted with IBM Key Protect, so it is safe to keep there.

**Note**: To learn how to get values encrypted using Key Protect, consult the [docs](https://cloud.ibm.com/docs/key-protect). You'll want to _wrap_ your secret using a [root key](https://cloud.ibm.com/docs/key-protect?topic=key-protect-envelope-encryption#key-types).

Our template simply creates a secret with this value like this:
```yaml
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: some-secret
stringData:
  clients-auth.json: |
    {
      "token": "[[ unwrap {{ .Values.auth_token | quote }} ]]"
    }
```

First, we see `{{ .Values.auth_token | quote }}` which picks up the encrypted `auth_token` value from the _values.yaml_ file and quotes it.

Second, we see `[[ unwrap ... ]]`. This is an _unwrap_ function, within different template delimiters (`[[]]`). The function is defined by this plugin and the delimiters are to make sure that only this plugin will process that function and return the true value.

Installing the chart via helm will create the following final manifest before sending it as an object to kubernetes:

```yaml
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: nats-clients-auth
stringData:
  clients-auth.json: |
    {
      "token": "password"
    }
```

Above you se can see the final result.

## Installation

```
$ go get -u github.ibm.com/renierm/helm-keyprotect
```

## Configuration

You need to have the following environment variables set:

* `IAM_TOKEN` - A JWT token from IBM Cloud IAM. A way to generate this using the IBM Cloud CLI is the following:
```
$ ibmcloud iam oauth-tokens # use the output value as the token 
```
* `SERVICE_INSTANCE_ID` - The GUID of your Key Protect service instance. You can use the IBM Cloud CLI to find it: `ibmcloud resource service-instances`
* `ROOT_KEY_ID` - The ID of the root key you will use. You find this on the Cloud UI.
* `KEYPROTECT_ENDPOINT` - Optional. Defaults to the us-south endpoint. All the endpoints are listed [here](https://cloud.ibm.com/docs/key-protect?topic=key-protect-regions#service-endpoints). Root keys do not exist in all endpoints, so the endpoint and root key you configure should correspond.

## Usage

If you have installed the plugin and have the above configuration set, then use the `--post-renderer` option in helm to use it:
```
helm upgrade -i --post-renderer helm-keyprotect ...
```

Questions or problems, please feel free to open an [issue](https://github.ibm.com/renierm/helm-keyprotect/issues/new).