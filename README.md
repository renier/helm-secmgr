# HELM Key Protect Postrenderer

With this helm v3 post-renderer plugin, you can have encrypted strings in your chart and they will be decrypted and made part of the manifest before it gets deployed to kubernetes.


## Example

Starting out from a helm chart top-level directory, let's say you have a _values.yaml_ file like this one:
```yaml
key_id: ba4d4437-1dca-4b7b-bca1-3effab898798
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
      "token": "[[ unwrap "{{ .Values.key_id }}" "{{ .Values.auth_token }}" ]]"
    }
```

We see a template function (`unwrap`) called with two parameters.

`.Values.key_id` is the ID of the root key in Key Protect being used to unwrap the encrypted value. You can find this id on your Key Protect instance dashboard on cloud.ibm.com.

`.Values.auth_token` is the encrypted value we want to unwrap (i.e. decrypt).

`[[ unwrap ... ]]` is the _unwrap_ function invocation, within different template delimiters (`[[]]`). The function is defined by this plugin and the delimiters are to make sure that only this plugin will process that function and return the true value.

Installing the chart via helm will create the following final manifest before sending it as an object to kubernetes:

```yaml
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: some-secret
stringData:
  clients-auth.json: |
    {
      "token": "password"
    }
```

Above you can see the final result.

### Limitation

Key Protect has a limit of 7Kb on the size of the plaintext you can wrap.

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
* `KEYPROTECT_ENDPOINT` - Optional. Defaults to the us-south endpoint. All the endpoints are listed [here](https://cloud.ibm.com/docs/key-protect?topic=key-protect-regions#service-endpoints). Root keys do not exist in all endpoints, so the endpoint and root key you configure should correspond.

## Usage

If you have installed the plugin and have the above configuration set, then use the `--post-renderer` option in helm to use it:
```
helm upgrade -i --post-renderer helm-keyprotect ...
```

Questions or problems, please feel free to open an [issue](https://github.ibm.com/renierm/helm-keyprotect/issues/new).
