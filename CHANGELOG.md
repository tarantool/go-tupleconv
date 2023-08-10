# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed 

- time formats, supported by `StringToDatetimeConverter`. Now there are two supported formats:
   - with numeric tz offset: `2006-01-02T15:04:05.999999999-0700`
   - with tz name: `2006-01-02T15:04:05.999999999 Europe/Moscow`