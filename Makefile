##
## Bureaucat — developer Makefile
##
## Modular: each concern lives in make/*.mk. Run `make` (or `make help`) for the
## full, self-documenting command list. The headline target is `make bootstrap` —
## one command that takes a fresh clone to a running, seeded local instance.
##

include make/00-config.mk
include make/10-dev.mk
include make/15-prod.mk
include make/20-setup.mk
include make/30-database.mk
include make/40-quality.mk
include make/50-clean.mk
include make/99-help.mk
