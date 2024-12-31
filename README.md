# Print to local printer from MQTT topic

---

## **Dependencies**

### **System Requirements**
- Only **Git**: required for cloning this repository.
    ```bash
    sudo apt update -y && sudo apt install git -y
    ```  
---

## **Installation**

1. Clone the repository, make scripts executable, run the initial setup script
   ```bash
   cd ~ && rm -rf printer && git clone https://github.com/jpfranca-br/ptiner.git && cd printer && chmod +x *.sh && ./setup.sh
   ```
2. After installing, use `manage.sh` for service management:
   ```bash
   ./manage.sh
   ```

---

## **Configuration**

Edit `printer.config` to match your setup:
```bash
MQTT_HOST = io.adafruit.com
MQTT_PORT = 8883
MQTT_USER = <user>
MQTT_PASS = <password like aio_****************************>
TOPIC = <user>/feeds/<topic_name>
TCP_HOST = localhost
TCP_PORT = 9100
```
