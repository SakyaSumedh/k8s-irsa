def function_handler(event, context):
    message = 'Hello {} {}!'.format(event['first_name'], event['last_name'])
    print(f"Message -> {message}") 
    return { 
        'message' : message
    }
