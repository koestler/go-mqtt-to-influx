# Changelog

## 2.10.0
- update various docs
- fix some typos in error messages
- fix bugs in pprof initialitation
- ttn: fix wrong unit V instead of °C for TempCDs

## 2.9.1
- TTN: Versions are only available for devices that are in the device registry. Add a matching by device name as fallback.
  - Replace ttn-fencyboy implementation with a check for "fencyboy" in the device name.
  - Add a matcher for d20s to dragino.
- Document fencyboy converter

## 2.9.0
- Add support for [Fencyboy](https://fencyboy.com/)
- Dependency bump
