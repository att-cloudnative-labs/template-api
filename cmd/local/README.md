# Local testing tool

Requires:
- `-options options.test.yaml` File with the configurations that will be passed 
as a map of options to the template.
- `wd <template_folder>` The folder that contains the `.genesis.yml` file and 
  the template files.
- `source <template_root>` The root folder of the template (i.e. base).

```bash
go build -o bin/gcli cmd/local/*.go
```

```bash
./bin/gcli -source base -wd="/mytemplate" -options "/mytemplate/options.test.yaml"
```

