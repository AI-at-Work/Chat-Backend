# Chat-Backend

This project simplifies the development of AI Agents by managing chat sessions. Chat-Backend, along with its peer services, provides chat summaries, file handling, and retrieval of specific numbers of chats from previous sessions to AI-Agents. It automatically manages chat responses, allowing developers to focus on implementing AI Agents rather than managing the infrastructure.

## System Flow Diagram

![System Flow Diagram](doc/flow.jpg)


## Services

- [Chat-AI](https://github.com/AI-at-Work/Chat-AI-Service)
- [Chat-UI](https://github.com/AI-at-Work/Chat-UI)
- [Sync-Backend](https://github.com/AI-at-Work/Sync-Backend)


  we need all of these services to make system work

## Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/AI-at-Work/Chat-Backend
   cd Chat-Backend
   ```

2. Copy the `.env.sample` to `.env` and configure the environment variables:
   ```bash
   cp .env.sample .env
   ```
   Edit the `.env` file to set your specific configurations and add openai api key.

3. Start the service:
   ```bash
   make && docker compose up -d --build
   ```
4. Create a user in the database and note down the `user_id` for further use:
   ```sql
   insert into user_data(username, models, balance) VALUES ('Test User', array[4,5,8], 1);
   ```
   `4` and `5` represent that user have access to `GPT4Turbo`, `GPT4Turbo09` and `llama3.1:8b` models.
   balance is in dollar.

5. Start `Sync-Backend` and `Chat-AI` service by following steps mention in respective repositories.

6. Restart the service:
   ```bash
   docker-compose down && docker-compose up -d
   ```
7. Start `Chat-UI` service for the existing UI by following steps mention in respective repository.

## Configuration

Key configuration options in the `.env` file:

- `SERVER_HOST` and `SERVER_PORT`: Host and port for the Chat-Backend server
- `DB_*`: PostgreSQL database configurations
- `REDIS_*`: Redis configurations
- `AI_SERVER_HOST` and `AI_SERVER_PORT`: AI service gRPC server details
- `MAX_FILE_SIZE`: Maximum allowed file upload size in MB
- `MAX_CHAT_HISTORY_CONTEXT`: Number of previous chat messages to include in context

Refer to the `.env.sample` file for a complete list of configuration options.

## API Documentation

### Model Id Mapping
	GPTTurbo125      = 0
	GPTTurbo         = 1
	GPTTurbo1106     = 2
	GPTTurboInstruct = 3
	GPT4Turbo        = 4
	GPT4Turbo09      = 5
	GPT4             = 6
	GPT40613         = 7
	llama3.1:8b      = 8


### gRPC Service

The Chat-Backend uses gRPC to communicate with the AI Agent. The protocol is defined in `proto/ai_service.proto`.

Key message types:
- `Request`: Contains user chat information, including user ID, session ID, chat message, model name, etc.
- `Response`: Contains the AI's response text and timestamp.

### WebSocket API

This documentation provides an overview of the WebSocket request handlers defined in the provided code. Each function generates a request to be sent via WebSocket for various operations related to user details, sessions, and chat messages. Below is the detailed explanation of each function and the corresponding message types.

## Table of Contents

- [Message Types](#message-types)
- [Functions](#functions)
  - [getUserDetails](#getuserdetails)
  - [getUserSessions](#getusersessions)
  - [getUserChatsBySessionId](#getuserchatsbysessionid)
  - [getUserChatsResponse](#getuserchatsresponse)
  - [deleteUserSession](#deleteusersession)
  - [modelList](#modellist)

## Message Types

The following constants represent the different message types used in the WebSocket requests:

- `MessageCodeUserDetails`: 0
- `MessageCodeListSessions`: 1
- `MessageCodeChatsBySessionId`: 2
- `MessageCodeChatMessage`: 3
- `MessageCodeSessionDelete`: 4
- `MessageCodeGetAIModels`: 5

## Functions

### getUserDetails

Generates a request to fetch user details.

#### Parameters

- `user_id` (String): The ID of the user.

```javascript
{
    type: MessageCodeUserDetails,
    data: {
        user_id: (String),
    },
}
```

#### Returns

```json
{
  "user_id": "String",
  "username": "String"
}
```

### getUserSessions

Generates a request to fetch the list of user sessions.

#### Parameters

- `user_id` (String): The ID of the user.

```javascript
{
    type: MessageCodeListSessions,
    data: {
        user_id: (String),
    },
}
```

#### Returns

```json
{
  "user_id": "String", 
  "session_info": [{
    "session_id": "String",
    "session_name": "String"
  }]
}
```

### getUserChatsBySessionId

Generates a request to fetch chats by session ID.

#### Parameters

- `user_id` (String): The ID of the user.
- `session_id` (String): The ID of the session.

```javascript
{
    type: MessageCodeChatsBySessionId,
    data: {
        user_id: (String),
        session_id: (String),
    },
}
```

#### Returns

```json
{
  "user_id": "String",
  "session_id": "String",
  "chat": "String"
}
```

### getUserChatsResponse

Generates a request to send a chat message in a specific session.

#### Parameters

- `user_id` (String): The ID of the user.
- `session_id` (String): The ID of the session.
- `model_id` (Int): The ID of the AI model.
- `message` (String): The chat message.
- `session_prompt` (String): The session prompt.
- `file_name` (String): The name of the file (optional).

```javascript
{
    type: MessageCodeChatMessage,
    data: {
        user_id: (String),
        session_id: (String),
        model_id: (Int),
        message: (String),
        session_prompt: (String),
        file_name: (String),
    },
}
```

#### Returns

```json
{
  "user_id": "String",
  "session_id": "String",
  "session_name": "String",
  "message": "String"
}
```

### deleteUserSession

Generates a request to delete a user session.

#### Parameters

- `user_id` (String): The ID of the user.
- `session_id` (String): The ID of the session.

```javascript
{
    type: MessageCodeSessionDelete,
    data: {
        user_id: userId,
        session_id: sessionId,
    },
}
```

#### Returns

```json
{
    "user_id": "String"
}
```

### modelList

Generates a request to fetch the list of AI models available.

#### Parameters

- `user_id` (String): The ID of the user.

```javascript
{
    type: MessageCodeGetAIModels,
    data: {
        user_id: userId,
    },
}
```

#### Returns

```json
{
    "models": ["String"]
}
```


## Contributing

We welcome contributions to the Chat-Backend project! Here's how you can contribute:

1. Fork the repository
2. Create a new branch (`git checkout -b feature/AmazingFeature`)
3. Make your changes
4. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
5. Push to the branch (`git push origin feature/AmazingFeature`)
6. Open a Pull Request

## License

This project is licensed under the [GNU GENERAL PUBLIC LICENSE V3](LICENSE).

## Contact

For any questions or suggestions, please open an issue in the GitHub repository.
