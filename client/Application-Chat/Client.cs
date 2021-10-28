using Jil;
using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Linq;
using System.Net.Security;
using System.Net.Sockets;
using System.Net;
using System.Security.Cryptography.X509Certificates;
using System.Text;
using System.Threading;
using System.Threading.Tasks;


namespace Application_Chat {
    public class Client {
        private TcpClient _tcpClient;
        private SslStream _encryptedStream; // Send & Receive TLS encrypted stream
        private Thread _threadReceive;
        private Thread _threadSend;
        public ManualResetEvent _exitEvent;  // Event to signal worker threads to terminate.
        private Timer keepalive; // kíp ờ lai
        // Identified by CommandID
        private Dictionary<string, ModelObject> _commandResults = new Dictionary<string, ModelObject>();
        // Identified by RoomID
        private Dictionary<string, List<ModelObject>> _receivedMessages = new Dictionary<string, List<ModelObject>>();
       
        private static Mutex rcvMutex = new Mutex();

        private ConcurrentQueue<string> _sendQueue = new ConcurrentQueue<string>();

        public class ModelObject {
            public struct Command {
                public string code;
                public string id;
                public string status;
                // public string key;
            }
            public struct User {
                public string name;
                public string alias;
                public string pass;
                public string ava;
                public string email;
            }
            public struct Room {
                public string id;
                public string name;
                public string type;
                public User[] members;
            }
            public struct Message {
                public string data;
                public string type;
                public string author;
                public string date;
                public Room[] existroom;
            }
            public Command command;
            public User user;
            public Room room;
            public Message message;

        }

        private struct _flag {
            public const int
                Login = 1,       //
                Register = 2,    //
                CreateRoom = 3,  //
                JoinRoom = 4,    //
                SendMessage = 5, //
                                 // 6 => Receive message
                GetAllUsers = 7, //
                                 // 8 => Polling
                Logout = 9,      // 
                LeaveRoom = 10;  //
        }
        
        public Client() {
            // C O N S T R U C T O R
            // Set up connection
            _tcpClient = new TcpClient();

            // Load-balancer's port || server's port (if using one server)
            _tcpClient.Connect("phmngcthanh.codes", 17749);
            
            Console.WriteLine("[+] TCP connected");

            _encryptedStream = new SslStream(
                _tcpClient.GetStream(),
                false 
              /*  ,(sender, cert, chain, err) => true   // Ignore certificate. Anything will do.*/
            );
            _encryptedStream.AuthenticateAsClient("phmngcthanh.codes");
            Console.WriteLine("[+] TLS negotiated.");

            // Kill switch, not signaled.
            _exitEvent = new ManualResetEvent(false);

            // Start 2 threads for network operations.
            _threadReceive = new Thread(ReceiveLoop) { IsBackground = true };
            _threadSend = new Thread(SendLoop) { IsBackground = true };
            _threadReceive.Start();
            _threadSend.Start();
           
        }

        public void closeProgram() {
          
            _exitEvent.Set();
        }
      //  public void closeProgram(object sender, FormClosingEventArgs e)
     //   {  }
        public List<string> getAllUser(ModelObject.User[] Members) {
            List<string> result = new List<string>();
            if (Members == null) {
                return result;
            }
            for (int i = 0; i < Members.Length; i++)
                result.Add(Members[i].name);
            return result;
        }

        public void _taskSend(ModelObject sendObj) {
            string jsonStr = JSON.Serialize(sendObj); // Null member => Omitted.
            Console.WriteLine(jsonStr);
            _sendQueue.Enqueue(jsonStr);
        }

        public void Send(ModelObject aBigObject) {
            //Pass data to a background task to avoid blocking main thread.
            Task.Run(() => _taskSend(aBigObject));
        }

        private void SendLoop() {
            string str;
            byte[] raw_data;
            while (!_exitEvent.WaitOne(0)) { // Check exit signal without blocking. 
                if (!_sendQueue.IsEmpty) {
                    if (_sendQueue.TryDequeue(out str)) {
                        str += "<EOF>";
                        raw_data = Encoding.UTF8.GetBytes(str);
                        Console.WriteLine(raw_data);
                        _encryptedStream.Write(raw_data, 0, raw_data.Length);
                        _encryptedStream.Flush(); // Make sure data is sent..?
                    }
                }
                else {
                    // If there is nothing to send, try polling (keepalive) connection?
                }
            }
            Console.WriteLine("[Client.cs] Send thread shutdown.");
        }
        //handling race condition brrrr 
        public List<ModelObject> GetMessage(string roomID) {
            rcvMutex.WaitOne();
            List<ModelObject> msg = new List<ModelObject>();

            if (_receivedMessages.ContainsKey(roomID) == true && _receivedMessages[roomID].Count > 0) {
                msg = _receivedMessages[roomID].ToList();
                _receivedMessages[roomID].Clear();
            }
            rcvMutex.ReleaseMutex();
            return msg;
        }

        public ModelObject GetCommandResult(string CommandID) {
            rcvMutex.WaitOne();
            ModelObject res = new ModelObject();
            if (_commandResults.ContainsKey(CommandID) == true) {
                res = _commandResults[CommandID];
                _commandResults.Remove(CommandID);
            }
            else {
                rcvMutex.ReleaseMutex();
                return null;
            }

            //Kiểm tra xem có CommandID có trong dictionary chưa - Không có sẽ báo bug
            rcvMutex.ReleaseMutex();
            return res;
        }

        private void ReceiveLoop() {

            byte[] buffer = new byte[2048];
            StringBuilder messageData = new StringBuilder();
            Decoder decoder = Encoding.UTF8.GetDecoder();
            int bytes=0;
            while (!_exitEvent.WaitOne(0)) { // Check exit signal without blocking. 
                // READ DATA
                do {
                    try
                    {
                        bytes = _encryptedStream.Read(buffer, 0, buffer.Length);
                        
                        char[] chars = new char[decoder.GetCharCount(buffer, 0, bytes)];
                        decoder.GetChars(buffer, 0, bytes, chars, 0);
                        messageData.Append(chars);
                        // Check for EOF.
                        if (messageData.ToString().IndexOf("<EOF>") != -1)
                        {
                            break;
                        }
                    }
                    catch
                    { }
                } while (bytes != 0);
                string response = messageData.ToString();  // JSON response string.
                messageData.Clear();


                // Process data.  
                if(response != "")
                {

                    response = response.Remove(response.Length - 5);
                    Console.WriteLine(response);
                    ModelObject recvObj = JSON.Deserialize<ModelObject>(response);
                    {
                        rcvMutex.WaitOne();
                        //int CommandType = Int32.Parse(recvObj.command.code);

                        if (recvObj.command.code == "6")
                        {  // Message
                            string roomID = recvObj.room.id;
                            if (!_receivedMessages.ContainsKey(roomID))
                            {
                                _receivedMessages[roomID] = new List<ModelObject>();
                            }
                            _receivedMessages[roomID].Add(recvObj);
                        }
                        else
                        {
                            //int CommandID = Int32.Parse(recvObj.command.ID);
                            Console.WriteLine("This is cmdID return:" + recvObj.command.id);
                            _commandResults[recvObj.command.id] = recvObj;
                        }
                        rcvMutex.ReleaseMutex();
                    }
                }
            }
            Console.WriteLine("[Client.cs] Receive thread shutdown.");
        }
    }
}
