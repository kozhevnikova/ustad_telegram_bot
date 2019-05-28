# ustad: a Telegram instant messenger bot 

### Objective: _extracting an audio from youtube video content._

The input data is a YouTube URL for a specific video; both mobile and desktop links are supported.  

For example: 
* https://www.youtube.com/watch?v=video_id
* https://m.youtube.com/watch?v=video_id

## Installation
Use *git clone* to copy the code from the remote repository to your laptop.  
Before starting the project, you need to fill the main.go file with the necessary credentials,  
otherwise the bot will be unable to launch.  

#### Third party library settings 

This project uses “channellogger” library from the repository at https://github.com/kozhevnikova/channellogger,  
which provides you with the possibility of sending log information to your Telegram channel.  

To apply this feature, enter Token and ChannelID data into the channel object on lines 53 and 54 on main() function,  
as in the example below.  

`channel = &channellogger.ChannelData{  
    Token: "your_bot_token_here",  
    ChannelID: "your_channel_id_here" 
 }`  

#### Bot settings

For successful execution of the bot, you need to fill the bot credentials with a token.  
Enter it in the botInitialization() function on line number 91.  

`err := os.Setenv("token", "your_token_here")`

If you do not know where to find a token, read https://core.telegram.org/bots#3-how-do-i-create-a-bot guide.  

## How the bot works
After starting the bot in a chat, the user can enter a link into the input field, which goes directly to the server.   
There, the server gets a message from the third-party library option “updates”  
(https://github.com/go-telegram-bot-api/telegram-bot-api) and processes the link to obtain a video id.  
If the link is not correct, the user receives an error message about the mistake.    
If it is correct, the server sends a request to get metadata for the requested video.    
Then, it downloads the video content if it is available in the local directory.   

NOTE: the video will be unavailable if it is forbidden for download. In such situations, the response is code 403.    

Next, the server processes the downloaded video to get the audio content by video id  
from the local directory and sends it to the user.  
When the user receives the audio file, the server deletes it from the directory. 

## Author
Jane Kozhevnikova - jane.kozhevnikova@gmail.com

## Status
Inactive

## License
MIT
