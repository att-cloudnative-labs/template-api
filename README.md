[![Go Report Card](https://goreportcard.com/badge/github.com/att-cloudnative-labs/template-api)](https://goreportcard.com/report/github.com/att-cloudnative-labs/template-api)
[![Build Status](https://travis-ci.org/att-cloudnative-labs/template-api.svg?branch=master)](https://travis-ci.org/att-cloudnative-labs/template-api)

# template-api
Parses .genesis.yml files and builds project from templates.

# Example cmd usage:
```bash
user$ template-api  --targetProjectKey KEY \
                    --targetRepoSlug myrepo \
                    --targetRepoFunctionalDomain domainkey \
                    --targetRepoProjectName test-genesis-proj \
                    --templateProjectName "My Project" \
                    --templateName "My Template" \
                    --options "option1=hello,option2=world"
```