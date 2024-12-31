#!/bin/bash

# Function to display the menu
show_menu() {
    echo -e "\n\nPrinter service management"
    echo "--------------------------"
    echo "1. View Log (Real-Time)"
    echo "2. View Service Status"
    echo "3. Enable Service"
    echo "4. (Re)start Service"
    echo "5. Stop Service"
    echo "6. Disable Service"
    echo "7. Exit"
    echo -n -e "\nEnter your choice [1-7]: "
}

# Menu loop
tput clear

while true; do
    show_menu
    read -r choice
    case $choice in
        1) echo "Displaying service log (real-time). Control+C to exit."
           tail /home/pi/printer/logs/service.log -f -n 40 || echo "Failed to view log." ;;
        2) echo "Displaying service status."
           sudo systemctl status printer.service || echo "Failed to get service status." ;;
        3) echo "Enabling service."
           sudo systemctl enable printer.service || echo "Failed to enable the service." ;;
        4) echo "(Re)starting service."
           sudo systemctl restart printer.service || echo "Failed to restart the service." ;;
        5) echo "Stopping service."
           sudo systemctl stop printer.service || echo "Failed to stop the service." ;;
        6) echo "Disabling service."
           sudo systemctl disable printer.service || echo "Failed to disable the service." ;;
        7) echo "Exiting the script. Goodbye!"
           exit 0 ;;
        *) echo "Invalid choice. Please select a valid option [1-7]." ;;
    esac
done
