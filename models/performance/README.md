---
title: Protobuf Conversion Performance
expires_at : never
tags: [diego-release, bbs]
---

# Protobuf Conversion Performance

Since BBS is changing from using [gogo/protobuf](https://github.com/gogo/protobuf) 
to [Google's Protobuf Implementation](https://google.golang.org/protobuf),
performance tests were added to verify the amount of change that an extra layer of conversion adds.

These tests use [gomega](https://github.com/onsi/gomega)'s `gmeasure` package to do 
benchmarking for the following scenarios using `models.DesiredLRP`:
 - Converting from a BBS model to a Protobuf binary format 
 - Converting from a BBS model to a Protobuf binary format and back again
 - Converting from a Protobuf binary format to a BBS model

The `DesiredLRP` model was chosen because it is the largest model that BBS uses, 
and therefore should be the slowest and most affected by any conversion changes.

# Updating Performance Results

> \[!NOTE\]
> If the new protobuf code has been merged into `develop`, the results should not need updating.

## Automatically

Use [diego-release](https://github.com/cloudfoundry/diego-release) `/scripts/update-bbs-proto-results.bash`

## Manually

`ginkgo --no-color . > ./results/results.txt`
