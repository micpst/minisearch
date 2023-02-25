# Changelog

## [1.1.0] 2023-02-26

### Added:
- Add search results ranking based on TF-IDF
- Add results pagination

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
