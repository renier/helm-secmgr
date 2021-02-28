# HELM Vault Postrenderer

With this helm v3 post-renderer plugin, you can have vault path references in your chart and they will be fetched and replaced in the  manifest before it gets deployed to kubernetes.

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
      "token": "[[ fetch "path/to/secret" "field_name" ]]"
    }
```

We see a template function (`fetch`) called with two parameters.

`[[ fetch ... ]]` is the _fetch_ function invocation, within different template delimiters (`[[]]`). The function is defined by this plugin and the delimiters are to make sure that only this plugin will process that function and return the true value.

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
$ go get -u github.ibm.com/renierm/helm-vault
```

## Configuration

You need to have the following environment variables set:

* `VAULT_ADDR` - URI to the vault endpoint you are using. 
* `VAULT_TOKEN` - The vault token used to authenticate.

## Usage

If you have installed the plugin and have the above configuration set, then use the `--post-renderer` option in helm to use it:
```
helm upgrade -i --post-renderer helm-vault ...
```

Questions or problems, please feel free to open an [issue](https://github.ibm.com/renierm/helm-vault/issues/new).
