#!/bin/bash

set -e  # Exit immediately if a command exits with a non-zero status

# Set repository details
REPO="jpfranca-br/printer"

# Fetch the latest release data using GitHub API
echo "Fetching latest release information..."
LATEST_RELEASE=$(curl -s https://api.github.com/repos/$REPO/releases/latest)

# Check if the API call was successful
if [[ $? -ne 0 || -z "$LATEST_RELEASE" ]]; then
    echo "Failed to fetch release information. Please check your network or repository details."
    exit 1
fi

# Extract the tag name and assets from the release
TAG_NAME=$(echo "$LATEST_RELEASE" | grep -Po '"tag_name": "\K.*?(?=")')
ASSETS=$(echo "$LATEST_RELEASE" | grep -Po '"browser_download_url": "\K.*?(?=")')

# Check if assets were found
if [[ -z "$ASSETS" ]]; then
    echo "No assets found in the latest release."
    exit 1
fi

# Display the available binaries
echo "Latest release: $TAG_NAME"
echo "Available binaries:"
ASSET_ARRAY=($ASSETS)
FILENAMES=()
for i in "${!ASSET_ARRAY[@]}"; do
    FILENAME=$(basename "${ASSET_ARRAY[$i]}")
    FILENAMES+=("$FILENAME")
    echo "$((i + 1)). $FILENAME"
done

# Prompt the user to select a binary
read -p "Enter the number of the binary to download: " SELECTION
if ! [[ "$SELECTION" =~ ^[0-9]+$ ]] || (( SELECTION < 1 || SELECTION > ${#ASSET_ARRAY[@]} )); then
    echo "Invalid selection. Exiting."
    exit 1
fi

# Download the selected binary
URL=${ASSET_ARRAY[$((SELECTION - 1))]}
echo "Downloading $URL..."
curl -L -o printer "$URL"

# Check if the download was successful
if [[ $? -eq 0 ]]; then
    chmod +x printer
    echo "Binary downloaded and saved as 'printer'."
else
    echo "Failed to download the binary."
    exit 1
fi

#############

SYSTEMD_PATH="/etc/systemd/system"

LOGROTATE_PATH="/etc/logrotate.d"

LOG_ROTATE_FILE="$LOGROTATE_PATH/printer-service"

CURRENT_DIR="$(pwd)"  # Get the current directory where the script is being executed

FILES=("manage.sh") # Files to be made executable

SERVICE_PATH="$SYSTEMD_PATH/printer.service"

SERVICE_FILE_CONTENT="[Unit]
Description=Printer Service
After=network.target

[Service]
Type=simple
ExecStart=$CURRENT_DIR/printer
Restart=always
RestartSec=5
User=$USER
WorkingDirectory=$CURRENT_DIR
StandardOutput=append:$CURRENT_DIR/logs/service.log
StandardError=append:$CURRENT_DIR/logs/service.log

[Install]
WantedBy=multi-user.target
"

LOG_ROTATE_CONTENT="$CURRENT_DIR/logs/service.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    copytruncate
    minsize 1
}
"

echo "==============================="
echo " Setting up Printer Service"
echo "==============================="

echo "Step 1: Creating service file in systemd directory..."
if echo "$SERVICE_FILE_CONTENT" | sudo tee "$SERVICE_PATH" > /dev/null; then
    echo "Service file created successfully."
else
    echo "Error creating service file!" >&2
    exit 1
fi

echo "Step 2: Creating logrotate configuration..."
if echo "$LOG_ROTATE_CONTENT" | sudo tee "$LOG_ROTATE_FILE" > /dev/null; then
    echo "Logrotate configuration created successfully."
else
    echo "Error creating logrotate configuration!" >&2
    exit 1
fi

echo "Step 3: Making script files executable..."
for file in "${FILES[@]}"; do
    if [ -f "$CURRENT_DIR/$file" ]; then
        chmod +x "$CURRENT_DIR/$file"
        echo "Made $file executable."
    else
        echo "Warning: $file not found, skipping..."
    fi
done

echo "Step 4: Creating log folder..."
if [ ! -d "$CURRENT_DIR/logs" ]; then
    if mkdir -p "$CURRENT_DIR/logs"; then
        echo "Log folder created."
    else
        echo "Error creating log folder!" >&2
        exit 1
    fi
else
    echo "Log folder already exists."
fi

echo "Step 5: Reloading systemd daemon..."
if sudo systemctl daemon-reload; then
    echo "Systemd daemon reloaded."
else
    echo "Error reloading systemd daemon!" >&2
    exit 1
fi

echo "Step 6: Enabling printer service..."
if sudo systemctl enable "printer.service"; then
    echo "Printer service enabled."
else
    echo "Error enabling printer service!" >&2
    exit 1
fi

echo "Step 7: Restarting printer service..."
if sudo systemctl restart "printer.service"; then
    echo "Printer service restarted successfully."
else
    echo "Error restarting printer service!" >&2
    exit 1
fi

echo "Step 8: Testing logrotate configuration..."
if sudo logrotate -f "$LOG_ROTATE_FILE"; then
    echo "Logrotate test completed successfully."
else
    echo "Error testing logrotate configuration!" >&2
    exit 1
fi

echo "==============================="
echo " Printer Service Setup Complete"
echo "==============================="
