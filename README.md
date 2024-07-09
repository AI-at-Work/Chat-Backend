# Chat-Backend
A Back-end for Chat Application 

# To start the service
`
make && docker-compose up -d --build
`

create the user in the database and use the generated user-id to make request.

# WebSocket Request Handlers

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

An object representing the request:

```javascript
{
  user_id: (String),
  username: (String),
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

An object representing the request:

```javascript
{
  user_id: (String), 
  session_info: [{
    session_id: (String),
    session_name: (String)
  }, ...],
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

An object representing the request:

```javascript
{
  user_id: (String),
  session_id: (String),
  chat: (String),
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

An object representing the request:

```javascript
{
  user_id: (String),
  session_id: (String),
  session_name: (String),
  message: (String),
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

An object representing the request:

```javascript
{
    user_id: userId,
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

An object representing the request:

```javascript
{
    models: [(String), ...]
}
```

This documentation provides an overview of how to use the provided functions to create requests for a WebSocket connection. Each function returns an object that can be sent through the WebSocket to perform the desired operation.