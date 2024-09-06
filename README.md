# FLIX indexer

Service for indexing `InteractionTemplate` data structures conforming to FLIX v1.1 standard, as defined in [FLIP 219](https://github.com/onflow/flips/blob/main/application/20230330-interaction-templates-1.1.0.md).

## REST API

### List all templates
```
GET /v1.1/templates
    => { data: InteractionTemplate[] }
```

### List template(s) by source code
```
GET /v1.1/templates?cadence_base64={base64 encoded cadence source code}
    => { data: InteractionTemplate[] }
```
