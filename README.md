# MiniSearch

[![build](https://github.com/micpst/minisearch/actions/workflows/build.yml/badge.svg)](https://github.com/micpst/minisearch/actions/workflows/build.yml)

Restful, in-memory, full-text search engine written in Go.

## ✅ Features

- [x] Full-text indexing of multiple fields in a document
- [x] Boolean queries with AND, OR operators between subqueries
- [ ] Exact phrase search
- [x] Document ranking based on TF-IDF
- [ ] Vector similarity search for semantic search
- [ ] Stemming-based query expansion for many languages
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
To search the index for documents that contain specific words, use the following request:
```bash
$ curl -G localhost:3000/api/v1/search \
    -d query=silicon%20brain \
    -d properties=title,abstract \
    -d bool_mode=AND
```

## 📄 License
All my code is MIT licensed. Libraries follow their respective licenses.
