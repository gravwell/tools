# tools
Collection of tools that do interesting things with Gravwell

To get started with Gravwell visit https://www.gravwell.io/community-edition

## Sentiment Training and Testing
A simple application that can generate a resource for the Gravwell sentiment search module to perform sentiment analysis using Bayes estimator

### Installation
```
go get -u github.com/gravwell/tools/sentiment/trainer
go get -u github.com/gravwell/tools/sentiment/tester
```

### Usage
You will need two data sets, one of positive sentiment and another of negative sentiment.  The IMDB review data is pretty good.

A copy of the training data can be found at https://github.com/cdipaolo/sentiment/tree/master/datasets/train

## Bro Named Fields

The bronamedfields tool walks a bro scripts directory and generates a resource for use in the namedfields search module.  Bro log entries are typically tab delimited and searching them using index offsets is a pain.  This tool allows you to generate a resource that will refer to bro data fields by name.

You will need to download a copy of the bro source tree that matches your running release.

### Installation
```go get -u github.com/gravwell/tools/brotools/namedfields```

### Usage
namedfields -o /tmp/namedfields.json /tmp/bro-2.5.4/scripts/
