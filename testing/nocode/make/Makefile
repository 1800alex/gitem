_bin_dir := $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/bin)

REPOS := $(shell $(_bin_dir)/gitm repos)

# No artifacts are produced
.PHONY: all show $(REPOS)

# Default target
all: $(REPOS)

# Rule to show the list of repositories
show:
	@echo $(REPOS)

# Rule to clone/pull a repository
$(REPOS):
	@$(_bin_dir)/gitm clone "$@"
