#!/bin/bash

# Function to display the menu
show_menu() {
    echo -e "\n\nPrinter service management"
    echo "--------------------------"
    echo "1. Reprint Order"
    echo "2. Test"
    echo "3. View Log (Real-Time)"
    echo "4. View Service Status"
    echo "5. Enable Service"
    echo "6. (Re)start Service"
    echo "7. Stop Service"
    echo "8. Disable Service"
    echo "9. Exit"
    echo -n -e "\nEnter your choice [1-9]: "
}

# Menu loop
tput clear

while true; do
    show_menu
    read -r choice
    case $choice in
        1)  while true; do
                echo -n -e "\nType order number to (re)print (0 to exit): "
                read -r pedido

                if [[ "$pedido" == "0" ]]; then
                    break
                fi

                if [[ -z "$pedido" ]]; then
                    echo "Order cannot be empty. Try again."
                    continue
                fi

                curl -X POST https://hook.us1.make.com/maqpsd6vicrtkvs2vml2ndsda33negtl \
                    -H "Content-Type: application/json" \
                    -H "token: 85ad8646-0363-40f9-ba25-6a78b4ed0218" \
                    -d "{\"pedido\": \"$pedido\"}"
            done ;;
        2)  echo -e "\nSending Order - OK - should be printed"
            curl -X POST https://hook.us1.make.com/maqpsd6vicrtkvs2vml2ndsda33negtl \
            -H "Content-Type: application/json" \
            -H "token: 85ad8646-0363-40f9-ba25-6a78b4ed0218" \
            -d '{"pedido": "1735322070"}'

            echo -e "\nSending order - non existant - should not be printed"
            curl -X POST https://hook.us1.make.com/maqpsd6vicrtkvs2vml2ndsda33negtl \
            -H "Content-Type: application/json" \
            -H "token: 85ad8646-0363-40f9-ba25-6a78b4ed0218" \
            -d '{"pedido": "1735322071"}'

            echo -e "\nSending order - not paid - should not be printed"
            curl -X POST https://hook.us1.make.com/maqpsd6vicrtkvs2vml2ndsda33negtl \
            -H "Content-Type: application/json" \
            -H "token: 85ad8646-0363-40f9-ba25-6a78b4ed0218" \
            -d '{"pedido": "1735318632"}'

            echo -e "\nSending order - with wrong token - should not be printed"
            curl -X POST https://hook.us1.make.com/maqpsd6vicrtkvs2vml2ndsda33negtl \
            -H "Content-Type: application/json" \
            -H "token: 85ad8646-0363-40f9" \
            -d '{"pedido": "1735318632"}'

            ;;
        3) echo "Displaying service log (real-time). Control+C to exit."
           tail /home/pi/printer/logs/service.log -f -n 40 || echo "Failed to view log." ;;
        4) echo "Displaying service status."
           sudo systemctl status printer.service || echo "Failed to get service status." ;;
        5) echo "Enabling service."
           sudo systemctl enable printer.service || echo "Failed to enable the service." ;;
        6) echo "(Re)starting service."
           sudo systemctl restart printer.service || echo "Failed to restart the service." ;;
        7) echo "Stopping service."
           sudo systemctl stop printer.service || echo "Failed to stop the service." ;;
        8) echo "Disabling service."
           sudo systemctl disable printer.service || echo "Failed to disable the service." ;;
        9) echo "Exiting the script. Goodbye!"
           exit 0 ;;
        *) echo "Invalid choice. Please select a valid option [1-9]." ;;
    esac
done
