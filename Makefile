BINARY_NAME=GitSpeak

DEST_DIR=$(HOME)/.GitSpeak/bin

all: install

build:
	go build -o $(BINARY_NAME)

install: build
	@mkdir -p $(DEST_DIR)
	@mv $(BINARY_NAME) $(DEST_DIR)
	@echo "Installation successful in $(DEST_DIR)"
	@echo "Ensure you add $(DEST_DIR) to your PATH if it's not already done."
	@echo 'You can do this by adding the following line to your profile file (.bashrc, .zshrc, etc.):'
	@echo 'export PATH=$$PATH:$(DEST_DIR)'

uninstall:
	@rm -f $(DEST_DIR)/$(BINARY_NAME)
	@echo "Uninstallation successful"

.PHONY: build install uninstall

