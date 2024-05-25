# MiniSearch

[![build](https://github.com/micpst/minisearch/actions/workflows/build.yml/badge.svg)](https://github.com/micpst/minisearch/actions/workflows/build.yml)

Restful, in-memory, full-text search engine written in Go.

## ✅ Features

- [x] Full-text indexing of multiple fields in a document
- [x] Exact phrase search
- [x] Document ranking based on BM25
- [x] Vector similarity search for semantic search
- [x] Stemming-based query expansion for many languages
- [x] Document deletion and updating with index garbage collection

## 🛠️ Installation

### Download binary
To download and run minisearch from a precompiled binary:
1. Download a precompiled version of minisearch from GitHub.
2. Run the server binary:
```bash
$ ./server
```

### Run with Docker
To run minisearch with Docker, use the **minisearch** Docker image:
```bash
$ docker run -d --name minisearch -p 3000:3000 micpst/minisearch:latest
```

### Build from source
To build and run minisearch from the source code:
1. Requirements: **go & make**
2. Install dependencies:
```bash
$ make setup
```
3. Build:
```bash
$ make build
```
4. Run the server binary:
```bash
$ ./bin/server
```

## 📘 Usage
### Add documents
Create a new document and add it to the index.
```bash
$ curl -X POST localhost:3000/api/v1/documents \
    -H 'Content-Type: application/json' \
    -d '{
      "title": "The Silicon Brain",
      "url": "https://micpst.com/posts/silicon-brain",
      "abstract": "The human brain is often described as complex..."
    }'
```

### Upload document dumps
Fill the index with a large number of documents at once by uploading a document dumps.
```bash
$ curl -X POST localhost:3000/api/v1/upload \
    -H 'Content-Type: multipart/form-data' \
    -F 'file[]=@/path/to/dataset1.xml.gz' \
    -F 'file[]=@/path/to/dataset2.xml.gz'
```
The dump should have the following structure:
```xml
<docs>
  <doc>
    <title>...</title>
    <url>...</url>
    <abstract>...</abstract>
  </doc>
  <doc>
    <title>...</title>
    <url>...</url>
    <abstract>...</abstract>
  </doc>
</docs>
```

### Update the document
Update the existing document and re-index it with the new fields.
```bash
$ curl -X PUT localhost:3000/api/v1/documents/<id> \
    -H 'Content-Type: application/json' \
    -d '{
      "title": "The Silicon Brain",
      "url": "https://micpst.com/posts/silicon-brain",
      "abstract": "The human brain is often described as complex..."
    }'
```

### Remove the document
Permanently delete the document and remove it from the index.
```bash
$ curl -X DELETE localhost:3000/api/v1/documents/<id>
```

### Search the index

#### Search properties
The `properties` parameter defines in which property to run our query.
```bash
$ curl -X POST localhost:3000/api/v1/search \
    -H 'Content-Type: application/json' \
    -d '{
      "query": "Brain",
      "properties": ["title"]
    }'
```
We are now searching for all the documents that contain the word `Brain` in the `title` property.

We can also search through nested properties:
```bash
$ curl -X POST localhost:3000/api/v1/search \
    -H 'Content-Type: application/json' \
    -d '{
      "query": "Mic",
      "properties": ["author.name"],
    }'
```
By default, MiniSearch searches in all searchable properties.

#### Exact match
The `exact` property finds all the document with an exact match of the `query` property.
```bash
$ curl -X POST localhost:3000/api/v1/search \
    -H 'Content-Type: application/json' \
    -d '{
      "query": "Brain",
      "properties": ["title"],
      "exact": true
    }'
```
We are now searching for all the documents that contain `exactly` the word `Brain` in the `title` property.

> Without the `exact` property, for example, the term `Brain-busting` would be returned as well, as it contains the word `Brain`.

#### Typo tolerance
The `tolerance` property allows specifying the maximum distance (following the Levenshtein algorithm) between the query and the searchable property.
```bash
$ curl -X POST localhost:3000/api/v1/search \
    -H 'Content-Type: application/json' \
    -d '{
      "query": "Brin",
      "properties": ["title"],
      "tolerance": 1
    }'
```
We are searching for all the documents that contain a term with an edit distance of `1` (e.g. `Brain`) in the `title` property.

> `tolerance` doesn't work together with the `exact` parameter. `exact` will have priority.

#### Pagination
The `offset` and `limit` properties allow paginating the results.
```bash
$ curl -X POST localhost:3000/api/v1/search \
    -H 'Content-Type: application/json' \
    -d '{
      "query": "Brain",
      "properties": ["title"],
      "offset": 10,
      "limit": 5
    }'
```
By default, MiniSearch limits the search results to 10, without any offset.

#### BM25 ranking
MiniSearch uses the BM25 algorithm to calculate the relevance of a document when searching.

You can edit the BM25 parameters by using the `relevance` property in the `search` configuration object.
```bash
$ curl -X POST localhost:3000/api/v1/search \
    -H 'Content-Type: application/json' \
    -d '{
      "query": "Brain",
      "properties": ["title"],
      "relevance": {
        // Term frequency saturation parameter.
        // Default value: 1.2
        // Recommended value: between 1.2 and 2
        "k": 1.2,

        // Length normalization parameter.
        // Default value: 0.75
        // Recommended value: > 0.75
        "b": 0.75,

        // Frequency normalization lower bound.
        // Default value: 0.5
        // Recommended value: between 0.5 and 1
        "d": 0.5
      }
    }'
```

## 📄 License
All my code is MIT licensed. Libraries follow their respective licenses.
