package plugin

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/fs185085781/v9os/internal/inface/distributed"
	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/pkg/util"
	"github.com/gorilla/websocket"
)

type WebSocketManager struct {
	upgrader   websocket.Upgrader
	clientMaps sync.Map
	mu         sync.Mutex
	syncMu     sync.Mutex
	syncTimer  *time.Timer
	log        logger.Logger
}

func (m *WebSocketManager) ReadPump(client *WSClient, handler IWebSocketHandler) {
	defer func() {
		m.CloseClient(client, handler)
	}()
	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("read error: %v", err)
			}
			break
		}
		// 将消息交给具体的业务处理器处理
		handler.OnMessage(client, message)
	}
}
func (m *WebSocketManager) GetClientMap(path string, needInit bool) (any, bool) {
	if needInit {
		return m.clientMaps.LoadOrStore(path, map[string]map[string]*WSClient{})
	} else {
		return m.clientMaps.Load(path)
	}
}
func (m *WebSocketManager) GetClientKeys() []string {
	var keys []string
	m.clientMaps.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	return keys
}
func (m *WebSocketManager) HasUser(userID string) bool {
	online := false
	m.clientMaps.Range(func(key, value interface{}) bool {
		userMap := value.(map[string]map[string]*WSClient)
		if _, ok := userMap[userID]; ok {
			online = true
			return false
		}
		return true
	})
	return online
}
func (m *WebSocketManager) WritePump(client *WSClient) {
	ticker := time.NewTicker(30 * time.Second) // 心跳
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			client.Conn.WriteMessage(websocket.TextMessage, message)
		case <-ticker.C:
			// 发送心跳
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (m *WebSocketManager) SendClient(client *WSClient, message []byte) bool {
	if client == nil {
		return false
	}
	defer func() {
		_ = recover()
	}()
	select {
	case client.Send <- message:
		return true
	default:
		return false
	}
}

func (m *WebSocketManager) AddClient(channelName string, client *WSClient) {
	func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		userMapAny, _ := m.GetClientMap(channelName, true)
		userMap := userMapAny.(map[string]map[string]*WSClient)
		cidMap := userMap[client.UserID]
		if cidMap == nil {
			userMap[client.UserID] = map[string]*WSClient{}
			cidMap = userMap[client.UserID]
		}
		cidMap[client.Cid] = client
	}()
	m.scheduleSyncWebsocketCloud()
}
func (m *WebSocketManager) CloseClient(client *WSClient, handler IWebSocketHandler) {
	func() {
		defer func() { _ = recover() }()
		close(client.Send)
	}()
	func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		a, ok := m.clientMaps.Load(handler.ChannelName())
		if ok {
			userMap := a.(map[string]map[string]*WSClient)
			cMap := userMap[client.UserID]
			if cMap != nil {
				delete(cMap, client.Cid)
				if len(cMap) == 0 {
					delete(userMap, client.UserID)
					if len(userMap) == 0 {
						m.clientMaps.Delete(handler.ChannelName())
					}
				}
			}
		}
	}()
	m.scheduleSyncWebsocketCloud()
	handler.OnClose(client)
	client.Conn.Close()
}

var wsManager *WebSocketManager

func GetWsManager() *WebSocketManager {
	//此处不用加锁,因为该代码首次执行一定来自go init触发,go init受单协程保护,参见Happens-Before
	if wsManager == nil {
		wsManager = &WebSocketManager{
			upgrader: websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			},
		}
		//该方法由运行时的导火线触发,这是Go最先调用的代码,因此必须延后执行
		ioc.Ioc().RegisterList(ioc.KeyAfterFunc, func() {
			wsManager.log = uioc.Log()
			ioc.Ioc().Register(ioc.KeyWebsocketUserResolver, wsManager.HasUser)
			wsManager.syncClearCloudPluginInfo()
			util.Go(wsManager.checkSyncWebsocketCloud)
			wsManager.log.Println("[Websocket管理器]已初始化")
		})
	}
	return wsManager
}

func (m *WebSocketManager) checkSyncWebsocketCloud() {
	for {
		time.Sleep(time.Minute)
		m.syncWebsocketCloud()
	}
}
func (m *WebSocketManager) Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return m.upgrader.Upgrade(w, r, nil)
}

// 用于首次初始化,清除所有map
func (m *WebSocketManager) syncClearCloudPluginInfo() {
	uioc.Get[distributed.DistributedProvider](ioc.KeyDistributedProvider).Websockets().SyncLocalUsers(nil)
}

func (m *WebSocketManager) syncWebsocketCloud() {
	userIDs := make([]string, 0)
	m.clientMaps.Range(func(key, value interface{}) bool {
		a, ok := m.clientMaps.Load(key)
		if ok {
			userMap := a.(map[string]map[string]*WSClient)
			for userID := range userMap {
				userIDs = append(userIDs, userID)
			}
		}
		return true
	})
	uioc.Get[distributed.DistributedProvider](ioc.KeyDistributedProvider).Websockets().SyncLocalUsers(userIDs)
}

func (m *WebSocketManager) scheduleSyncWebsocketCloud() {
	m.syncMu.Lock()
	defer m.syncMu.Unlock()
	if m.syncTimer != nil {
		m.syncTimer.Reset(2 * time.Second)
		return
	}
	m.syncTimer = time.AfterFunc(2*time.Second, func() {
		m.syncMu.Lock()
		m.syncTimer = nil
		m.syncMu.Unlock()
		m.syncWebsocketCloud()
	})
}

type IWebSocketHandler interface {
	ChannelName() string
	// OnOpen 连接建立时的钩子函数
	OnOpen(client *WSClient)
	// OnMessage 处理客户端消息
	OnMessage(client *WSClient, message []byte)
	// OnClose 连接关闭时的钩子函数
	OnClose(client *WSClient)
}
type WSClient struct {
	Conn   *websocket.Conn
	UserID string
	Send   chan []byte
	Cid    string
}

// 遵守该格式的消息才支持分布式事件转发
type WebsocketMessage struct {
	ID       string    `json:"id"`
	From     string    `json:"from"`
	To       string    `json:"to"`
	Msg      string    `json:"msg"`
	Type     string    `json:"type"`
	Status   string    `json:"status"` //success 成功 fail 失败
	DateTime time.Time `json:"date_time"`
}

type WebsocketMessageHandler func(msg WebsocketMessage) bool
