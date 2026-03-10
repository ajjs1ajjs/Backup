package storage 
type Engine struct{} 
func NewEngine(s string)*Engine{return new(Engine)} 
func(e*Engine)StoreChunk(h string,d[]byte)(string,error){return h,nil} 
func(e*Engine)GetChunk(h string)([]byte,error){return nil,nil} 
func(e*Engine)GetStorageStats()*StorageStats{return new(StorageStats)} 
type StorageStats struct{} 
