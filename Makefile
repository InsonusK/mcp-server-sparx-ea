init:
	python3 -m venv .venv
	.venv/bin/pip install -r requirements.txt
	.venv/bin/aism sync

ai-skill-sync:
	.venv/bin/aism sync