package spider

// 安全发送
func SafeSendTorrent(ch chan *Torrent, t *Torrent)(closed bool){
	defer func() {
		if v := recover(); v != nil{
			closed = true
		}
	}()
	if t == nil{
		return
	}
	ch <- t
	return false
}