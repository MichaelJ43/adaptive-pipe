.PHONY: up down test test-integration ui e2e

up:
	docker compose up --build

down:
	docker compose down

# Go unit tests (requires Go 1.22+ on PATH)
test:
	cd src && go test ./... -count=1

# Requires: docker compose up -d
test-integration:
	cd src && INTEGRATION=1 go test -tags=integration ./integration/... -count=1 -v

ui:
	cd src/ui && npm install && npm run dev

e2e:
	cd src/ui && npm install && npx playwright install && npm run test:e2e
