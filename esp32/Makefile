FQBN:=esp32:esp32:esp32 # board name
PORT:=/dev/ttyUSB0 # port used for uploading
PROJECT:=$(notdir $(PWD)) # project name
BAUD:=115200 # baudrate for the serial monitor (if used)

help:
	@echo "Usage:"
	@echo "  make [command]"
	@echo ""
	@echo "Commands:"
	@echo "  help - Show this help message"
	@echo "  compile - Compile the sketch"
	@echo "  upload - Upload the sketch to the board"
	@echo "  monitor - Start the serial monitor"
	@echo "  all - Compile, upload and start the serial monitor"

compile:
	@echo "Compiling $(PROJECT) for $(FQBN)"
	arduino-cli $@ --fqbn $(FQBN) .

upload:
	@echo "Uploading $(PROJECT) to $(FQBN)"
	arduino-cli $@ -p $(PORT) --fqbn $(FQBN) .

monitor:
	@echo "Monitoring $(PROJECT) on $(FQBN)"
	arduino-cli $@ -p $(PORT) --fqbn $(FQBN) --config baudrate=$(BAUD)

all: compile upload monitor

