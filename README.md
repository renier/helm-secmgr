# HELM Postrenderer for IBM Secrets Manager

With this helm v3 post-renderer plugin, you can have vault (via [Secrets Manager](https://cloud.ibm.com/docs/secrets-manager)) path references in your chart and they will be fetched and replaced in the manifest before it gets deployed to kubernetes.

This tool only supports _arbitrary_ secrets per the [Secrets Manager API](https://cloud.ibm.com/apidocs/secrets-manager).

## Example

In the yaml template, use the post-renderer function with special delimiters, and passing the vault path and field name that you want to swap in.:
```yaml
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: some-secret
stringData:
  clients-auth.json: |
    {
      "token": "[[ secret_ref "groups/abcd-b6ee-defb-d166-39f9999000d4/2d87d1cc-21da-431c-90f5-2bbf359126dc" ]]"
    }
```

We see a template function (`fetch`) called with two parameters.

`[[ secret_ref ... ]]` is the secret fetching function invocation within different template delimiters (`[[]]`). The function is defined by this plugin and the delimiters are to make sure that only this plugin will process that function and return the true value.

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

## Installation

```
$ go get -u github.com/renier/helm-secmgr
```

## Configuration

You need to have the following environment variables set:

* `VAULT_ADDR` - URI to the vault endpoint you are using. You can get this endpoint from your Secrets Manager instance.
* `VAULT_TOKEN` - The vault token used to authenticate. [How to get this?](https://cloud.ibm.com/docs/secrets-manager?topic=secrets-manager-configure-vault-cli). It will be read from `~/.vault-token` if it is not set in the environment.

## Usage

If you have installed the plugin and have the above configuration set, then use the `--post-renderer` option in helm to use it:
```
helm upgrade -i --post-renderer helm-secmgr ...
```

Questions or problems, please feel free to open an [issue](https://github.com/renier/helm-secmgr/issues/new).
