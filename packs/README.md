# Built-in Packs

This directory contains built-in Draft packs.

Packs in this directory are automatically packaged into the application. The
format of this directory is:

```
packs/
  |
  |- PACKNAME
  |     |
  |     |- chart/
  |     |    |- Chart.yaml
  |     |    |- ...
  |     |- Dockerfile
  |     |- detect
  |     |- ...
  |
  |- PACK2
        |-...
```

Packs are bundled using `make generate`, which will create a binary representation
of the chart and store that in `pkg/draft/pack/generated`.
