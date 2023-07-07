# Message Queueing System

This repository contains the source code for a Message Queueing System built using Golang. The system is responsible for consuming messages from a Kafka topic, processes the productID and downloads the images from the web and compress them and finally save it to a local folder.

## Prerequisites

Before running the Message Queueing System, make sure you have the following prerequisites installed on your system:

- Go programming language (go1.20.4)
- Kafka (3.4.0)
- PostgreSQL

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/ankitchahal20/message-queueing-system.git
   ```

2. Navigate to the project directory:

   ```bash
   cd message-queueing-system
   ```

3. Install the required dependencies:

   ```bash
   go mod tidy
   ```

4. install kafka and its dependencies
    ```
        brew install kafka
        zookeeper-server-start /usr/local/etc/kafka/zookeeper.properties
        kafka-server-start /usr/local/etc/kafka/server.properties
        cd /usr/local/etc/kafka
        ./kafka-topics.sh --create --topic my-kafka-topic --bootstrap-server localhost:9092 --partitions 3 --replication-factor 2
        ./kafka-topics.sh --describe --topic my-kafka-topic --bootstrap-server localhost:9092
    ```
4. DB setup
    ```
    Use the scripts inside sql-scripts directory to create the tables in your db.
    ```
5. Defaults.toml
Add the values to defaults.toml and execute `go run main.go` from the cmd directory.

curl -i -k -X POST \
  http://127.0.0.1:8080/v1/product/create \
  -H "transaction-id: 288a59c1-b826-42f7-a3cd-bf2911a5c351" \
  -H "content-type: application/json" \
  -d '{
  "user_id": 11,
  "product_name": "ANC17",
  "product_description": "Nice Project",
  "product_images": ["https://cdn.pixabay.com/photo/2013/10/15/09/12/flower-195893_150.jpg","https://images.pexels.com/photos/2014422/pexels-photo-2014422.jpeg","https://images.pexels.com/photos/2014421/pexels-photo-2014421.jpeg"],
  "product_price": 10
}'

## Project Structure

The project follows a standard Go project structure:

- `cmd/`: Contains the main entry points for the application.
   - `Images/`: Stores the compressed images locally
- `config/`: Configuration file for the application.
- `internal/`: Contains the internal packages and modules of the application.
  - `config/`: Global configuration which can be used anywhere in the application.
  - `constants/`: Contains constant values used throughout the application.
  - `db/`: Contains the database package for interacting with PostgreSQL.
  - `kafka/`: Contains the Kafka package for consuming and producing messages.
  - `middleware`: Contains the logic to validate the incoming request
  - `models/`: Contains the data models used in the application.
  - `producterror`: Defines the errors in the application
  - `service/`: Contains the business logic and services of the application.
  - `server/`: Contains the server logic of the application.
  - `utils/`: Contains utility functions and helpers.
- `main.go`: Main entry point of the application.
- `README.md`: This file.

## Contributing

Contributions to the Message Queueing System are welcome. If you find any issues or have suggestions for improvement, feel free to open an issue or submit a pull request.

## License

The Message Queueing System is open-source and released under the [MIT License](LICENSE).

## Contact

For any inquiries or questions, please contact:

- Ankit Chahal
- ankitchahal20@gmail.com

Feel free to reach out with any feedback or concerns.
