pkg - go utility packages
=========================

A collection of utility packages that I use for my go projects.

## Packages

* container – provides a set of container presets for testing. For instance, you can create a Postgres container for
  data layer tests.
* i18n – contains a `Localizer` that can load localisation data and localize the text with options.
* net – utility functions for working with network.
* postgres – a wrapper around `sql.DB` that uses pgx drive under the hood. There are also some helpful utility methods
  to work with Postgres.
