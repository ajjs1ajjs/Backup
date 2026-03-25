SHELL := /bin/bash

install:
	@echo "Creating virtualenv and installing in editable mode..."
	@python3 -m venv venv
	@source venv/bin/activate; pip install -e .

test:
	@pytest -q

list:
	@source venv/bin/activate; python -m novabackup list-vms

normalize:
	@source venv/bin/activate; python -m novabackup normalize $(vm_type)

.PHONY: install test list normalize
