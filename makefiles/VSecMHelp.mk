# /*
# |    Protect your secrets, protect your sensitive data.
# :    Explore VMware Secrets Manager docs at https://vsecm.com/
# </
# <>/  keep your secrets… secret
# >/
# <>/' Copyright 2023–present VMware Secrets Manager contributors.
# >/'  SPDX-License-Identifier: BSD-2-Clause
# */

help:
	@echo "--------------------------------------------------------------------"
	@echo "          🛡️ VMware Secrets Manager: Keep your secrets… secret."
	@echo "          🛡️ https://vsecm.com/"
	@echo "--------------------------------------------------------------------"
	@echo "        ℹ️ This Makefile assumes you use Minikube and Docker"
	@echo "        ℹ️ for most operations."
	@echo "--------------------------------------------------------------------"
	@echo "If you are on the build server, stop the cronjob first: \`crontab -e\`"
	@echo "Or \`sudo service cron stop\`"
	@echo "--------------------------------------------------------------------"

	@if [ "`uname`" = "Darwin" ]; then \
		if type docker > /dev/null 2>&1; then \
			echo "  Using Docker for Mac?"; \
			echo "        ➡ 'make mac-tunnel' to proxy to the internal registry."; \
		else \
			echo "  Docker is not installed on this Mac."; \
		fi; \
	fi

	@echo ""

	@if [ -z "$(DOCKER_HOST)" -o -z "$(MINIKUBE_ACTIVE_DOCKERD)" ]; then \
		echo "  Using Minikube? If DOCKER_HOST and MINIKUBE_ACTIVE_DOCKERD are"; \
		echo '  not set, then run: eval $$(minikube -p minikube docker-env)'; \
		echo "        ➡ \$$DOCKER_HOST            : ${DOCKER_HOST}"; \
		echo "        ➡ \$$MINIKUBE_ACTIVE_DOCKERD: ${MINIKUBE_ACTIVE_DOCKERD}"; \
	else \
	    echo "  Make sure DOCKER_HOST and MINIKUBE_ACTIVE_DOCKERD are current:"; \
		echo '          eval $$(minikube -p minikube docker-env)'; \
	    echo "          (they may change if you reinstall Minikube)"; \
		echo "        ➡ \$$DOCKER_HOST            : ${DOCKER_HOST}"; \
		echo "        ➡ \$$MINIKUBE_ACTIVE_DOCKERD: ${MINIKUBE_ACTIVE_DOCKERD}"; \
	fi

	@echo "--------------------------------------------------------------------"
	@echo "  Prep/Cleanup:"
	@echo "          ˃ make k8s-delete;make k8s-start;"
	@echo "          ˃ make clean;"
	@echo "   ⦿ Optional Minikube Settings:"
	@echo "          > make k8s-start CPU=4 MEMORY=4096"
	@echo "          or"
	@echo "          > make k8s-start VSECM_MINIKUBE_CPU_COUNT=4 VSECM_MINIKUBE_MEMORY=4096"
	@echo "   ⦿ Generate Proto files:"
	@echo "          ˃ make generate-proto-files"
	@echo "--------------------------------------------------------------------"
	@echo "  Installation:"
	@echo "    ⦿ Distroless images:"
	@echo "          ˃ make deploy;make test;"
	@echo "    ⦿ Distroless FIPS images:"
	@echo "          ˃ make deploy-fips;make test;"
	@echo "    ⦿ Distroless images:"
	@echo "          ˃ make deploy-photon;make test;"
	@echo "    ⦿ Photon FIPS images:"
	@echo "          ˃ make deploy-photon-fips;make test;"
	@echo "--------------------------------------------------------------------"
	@echo "  Example Use Cases:"
	@echo "    Using local images:"
	@echo "          ˃ make example-sidecar-deploy-local;"
	@echo "          ˃ make example-sdk-deploy-local;"
	@echo "          ˃ make example-multiple-secrets-deploy-local;"
	@echo "    Using remote images:"
	@echo "          ˃ make example-sidecar-deploy;"
	@echo "          ˃ make example-sdk-deploy;"
	@echo "          ˃ make example-multiple-secrets-deploy;"

h:
	@echo "➡ 'make mac-tunnel'";
	@echo "˃ make k8s-delete;make k8s-start;"
	@echo '⦿ eval $$(minikube -p minikube docker-env)';
	@echo "˃ make clean;"
	@echo "˃ make generate-proto-files"
	@echo "˃ make build-local;make deploy-local;make test-local;"
	@echo "˃ make build;make deploy;make test;"
	@echo "˃ make tag;"
	@echo "--------------------------------------------------------------------"
	@echo "If you are on the build server, stop the cronjob first: \`crontab -e\`"
	@echo "Or \`sudo service cron stop\`"
	@echo "--------------------------------------------------------------------"
	@echo "˃ make build-local;make deploy-local;make test-local;"
	@echo "˃ make build-local;make deploy-fips-local;make test-local;"
	@echo "˃ make build-local;make deploy-photon-local;make test-local;"
	@echo "˃ make build-local;make deploy-photon-fips-local;make test-local;"
	@echo "--------------------------------------------------------------------"
	@echo "Building and Remote Testing:"
	@echo "    > Make sure you 'export DOCKER_CONTENT_TRUST=1' first."
	@echo "    > --"
	@echo "    > make build;make deployABC; make test"
	@echo "    > (where ABC is one of: -fips, -photon, -photon-fips)"
	@echo "Tagging:"
	@echo "    ˃ make tag;"
	@echo "--------------------------------------------------------------------"
