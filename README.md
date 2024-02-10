## Mongo-Cmp

Mongo-Cmp is a simple tool to compare all collections between two MongoDB clusters.

## Usage

```shell
mongo-cmp --from "mongodb://localhost:27017" --to "mongodb://localhost:27018"
```

## Installation

```shell
brew tap fadyat/apps
brew install fadyat/apps/mongo-cmp
```

```shell
# Here is an example of how to install the app on a M1 Mac
# For different architectures, please look at the releases page
# and download the correct version for your architecture
# https://github.com/fadyat/mongo-cmp/releases

curl -L https://github.com/fadyat/mongo-cmp/releases/download/v0.0.1/mongo-cmp_Darwin_arm64.tar.gz > mongo-cmp.tar.gz
tar -xvf mongo-cmp.tar.gz
mv mongo-cmp ~/bin
rm mongo-cmp.tar.gz
```
