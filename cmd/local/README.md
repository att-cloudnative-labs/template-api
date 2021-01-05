# Local testing tool

Requires:
- `-options options.test.yaml` File with the configurations that will be passed 
as a map of options to the template.
- `wd <template_folder>` The folder that contains the `.genesis.yml` file and 
  the template files.
  
```bash
go build -o bin/gcli cmd/local/*.go
```

```bash
./bin/gcli -wd="/mytemplate" -options "/mytemplate/options.test.yaml"
```

Example of `options.test.yaml`

```yaml
# The folder that contains the template files
source: base
# The name/key used to identify the project template
template_name: Go SDK Archectype

# Key-value pairs send to be replaced in the template
settings:
  project_name: Demo
  project_url: http://someurl.com
```