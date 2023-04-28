# Changelog

## [1.2.0] 2023-04-28

### Added:
- Search results ranking based on BM25
- Stemming-based query expansion for many languages
- Vector similarity search for semantic search

### Changed:
- Change search API endpoint HTTP method to POST
- Move search params from query string to request body

## [1.1.0] 2023-02-26

### Added:
- Search results ranking based on TF-IDF
- Results pagination

### Changed:
- Rename project to `minisearch`

## [1.0.1] 2023-02-24

### Changed:
- Bump go version to 1.20
- Update dependencies
- Improve overall performance

### Removed:
- Remove `github.com/cornelk/hashmap` dependency

## [1.0.0] 2023-02-13

### Added:
- Full-text indexing of multiple fields in a document
- Boolean queries with AND, OR operators between subqueries
- Document deletion and updating with index garbage collection 
