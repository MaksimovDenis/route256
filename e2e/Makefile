.PHONY: test-e2e
test-e2e: ## Запуск e2e (integration) тестов
	go test -v -tags=integration -count=1 -timeout=5m ./...

generate-allure: ## Генерация и открытие отчёта Allure
	allure generate ./allure-results --clean -o allure-report
	allure open allure-report
