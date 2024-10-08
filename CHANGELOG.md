# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Bump `go-tarantool` dependency to `v2` (#6).

## [0.1.0] - 2023-09-08

### Changed

- Time formats, supported by `StringToDatetimeConverter`. Now there are two supported formats:
   - With numeric tz offset: `2006-01-02T15:04:05.999999999-0700`
   - With tz name: `2006-01-02T15:04:05.999999999 Europe/Moscow`

### Added

- `DatetimeToStringConverter`: converter from `datetime.Datetime` to string with a format
  similar to that supported by `StringToDatetimeConverter`.
- `IntervalToStringConverter`: converter from `datetime.Interval` to string with a format
  similar to that supported by `StringToIntervalConverter`.
