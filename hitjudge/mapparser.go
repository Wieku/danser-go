package hitjudge

import (
	"danser/beatmap"
	"danser/beatmap/objects"
	"danser/bmath"
	"danser/replay"
	"danser/settings"
	"flag"
	"github.com/Mempler/rplpa"
	"log"
	"os"
)

var None = rplpa.KeyPressed{
	LeftClick:  false,
	RightClick: false,
	Key1:       false,
	Key2:       false,
}

func ParseMap(filename string) *beatmap.BeatMap{
	file, err := os.Open(filename)
	if err == nil {
		defer file.Close()
		beatMap := beatmap.ParseBeatMap(file)
		beatmap.ParseObjectsbyPath(beatMap, filename)
		return beatMap
	}else{
		panic(err)
	}
}

func ParseReplay(name string) *rplpa.Replay {
	return replay.ExtractReplay(name)
}

func ParseHits(mapname string, replayname string){
	settingsVersion := flag.Int("settings", 0, "")
	settings.LoadSettings(*settingsVersion)

	// 加载map
	b := ParseMap(mapname)
	convert_CS := 32 * (1 - 0.7 * (b.CircleSize - 5) / 5)
	// 加载replay
	r := ParseReplay(replayname).ReplayData

	// 依次处理HitObject
	keyindex := 3
	time := r[1].Time + r[2].Time
	//for _, obj := range b.HitObjects {
	//	if obj != nil {
	//		if o, ok := obj.(*objects.Slider); ok {
	//			log.Println(o.TickPoints)
	//		}
	//		if o, ok := obj.(*objects.Circle); ok {
	//			log.Println(o)
	//		}
	//	}
	//}
	for k := 0; k < 20; k++ {
		obj :=  b.HitObjects[k]
		if obj != nil {
			// 滑条
			if o, ok := obj.(*objects.Slider); ok {
				log.Println("Slider", o.GetBasicData().StartPos, o.GetBasicData().EndPos, o.TickPoints)
			}
			// note
			if o, ok := obj.(*objects.Circle); ok {
				// 寻找最近的Key
				isfind, nearestindex, lasttime := findNearestKey(keyindex, time, r, o.GetBasicData().StartTime, o.GetBasicData().StartPos, b.ARms, b.ODMiss, convert_CS)
				if isfind {
					// 如果找到，判断hit结果，设置下一个index+1
					keyhitresult := judgeHitResult(nearestindex, lasttime, r, o.GetBasicData().StartTime, b.ODMiss, b.OD300, b.OD100, b.OD50)
					switch keyhitresult {
					case Hit300:
						log.Println("Circle", o.GetBasicData().StartPos, o.GetBasicData().StartTime, "300")
						break
					case Hit100:
						log.Println("Circle", o.GetBasicData().StartPos, o.GetBasicData().StartTime, "100")
						break
					case Hit50:
						log.Println("Circle", o.GetBasicData().StartPos, o.GetBasicData().StartTime, "50")
						break
					case HitMiss:
						log.Println("Circle", o.GetBasicData().StartPos, o.GetBasicData().StartTime, "Miss")
						break
					}
					time = lasttime + r[nearestindex].Time
					log.Println("hit in", time)
					// 寻找状态改变后的时间点
					keyindex, time = findRelease(nearestindex, time, r)
					time -= r[keyindex].Time
				}else {
					// 如果没找到，输出miss，设置下一个index
					log.Println("Circle", o.GetBasicData().StartPos, o.GetBasicData().StartTime, "Miss")
					keyindex = nearestindex
					time = lasttime
				}
			}
		}
	}
}

// 定位Key放下的位置
func findRelease(keyindex int, starttime int64, r []*rplpa.ReplayData) (int, int64) {
	keypress := r[keyindex].KeyPressed
	index := keyindex
	time := starttime
	for {
		index++
		time += r[index].Time
		// 如果按键状态改变，则返回
		if isPressChanged(*keypress, *r[index].KeyPressed) {
			log.Println("Find release before", time)
			return index, time
		}
	}
}

// 确定是否出现按下状态的改变
func isPressChanged(p1 rplpa.KeyPressed, p2 rplpa.KeyPressed) bool {
	if p1!=p2 {
		// 如果不相等
		if p2==None{
			// 如果没有按键，则肯定状态改变
			return true
		}else {
			// 否则，如果p2按下了某个键，p1必须也按下了这个键，否则状态改变
			if p2.Key1{
				if !p1.Key1{
					return true
				}
			}
			if p2.Key2{
				if !p1.Key2{
					return true
				}
			}
			if p2.LeftClick{
				if !p1.LeftClick{
					return true
				}
			}
			if p2.RightClick{
				if !p1.RightClick{
					return true
				}
			}
			return false
		}
	}else {
		//相等，无改变
		return false
	}
}

// 寻找最近的击中的Key
func findNearestKey(start int, starttime int64, r []*rplpa.ReplayData, requirehittime int64, requirepos bmath.Vector2d, ARms float64, ODMiss float64, CS float64) (bool, int, int64) {
	index := start
	time := starttime
	for {
		hit := r[index]
		//log.Println("Find move", hit, hit.Time + time, isInCircle(hit, requirepos, CS), isPressed(hit), bmath.NewVec2d(float64(hit.MosueX), float64(hit.MouseY)), requirepos, bmath.Vector2d.Dst(bmath.NewVec2d(float64(hit.MosueX), float64(hit.MouseY)), requirepos), CS)
		//if hit.Time + time > 19200 {
		//	os.Exit(2)
		//}
		// 判断是否在圈内
		if isInCircle(hit, requirepos, CS){
			// 如果在圈内，且按下按键
			if isPressed(hit) {
				realhittime := hit.Time + time
				// 判断这个时间点和object时间点的关系
				//log.Println("Judge", realhittime, requirehittime, ODMiss)
				if isHitOver(realhittime, requirehittime, ODMiss) {
					// 如果已经超过这个object的最后hit时间，则未找到最接近的Key，直接返回这个时间点
					//log.Println("isHitOver")
					return false, index, time
				}else if isHitMiss(realhittime, requirehittime, ODMiss){
					// 如果落在这个object的区域内，则找到Key，返回这个Key的时间点
					//log.Println("isHitMiss")
					return true, index, time
				}else {
					// 如果在区域之前，继续找
					index++
					time += hit.Time
				}
			}else{
				// 没按下，继续
				index++
				time += hit.Time
			}
		}else {
			// 如果不在圈内，且按下按键
			if isPressed(hit) {
				realhittime :=hit.Time
				// 判断这个时间点和object时间点的关系
				if realhittime > requirehittime {
					// 如果在缩圈结束之后按下，没效果，等于没找到，返回这个Key的时间点
					return false, index, time
				}else if float64(realhittime) >= float64(requirehittime) - ARms {
					// 如果在出现note到缩圈之前按下，miss，等于没找到，返回下一个Key的时间点
					return false, index + 1, time
				}else {
					// 如果在note出现之前，继续找
					index++
					time += hit.Time
				}
			}else{
				// 没按下，继续
				index++
				time += hit.Time
			}
		}
	}
}

// 该时间点是否按下按键
func isPressed(hit *rplpa.ReplayData) bool {
	press := hit.KeyPressed
	return press.LeftClick || press.RightClick || press.Key1 || press.Key2
}

func isInCircle(hit *rplpa.ReplayData, requirepos bmath.Vector2d, CS float64) bool {
	realpos := bmath.NewVec2d(float64(hit.MosueX), float64(hit.MouseY))
	return bmath.Vector2d.Dst(realpos, requirepos) <= CS
}

// 是否超过object的最后时间点
func isHitOver(realhittime int64, requirehittime int64, ODMiss float64) bool {
	return float64(realhittime) > float64(requirehittime) + ODMiss
}

// 判断hit结果
func judgeHitResult(index int, lasttime int64, r []*rplpa.ReplayData, requirehittime int64, ODMiss float64, OD300 float64, OD100 float64, OD50 float64) HitResult{
	realhittime := r[index].Time + lasttime
	log.Println("Judge Hit", realhittime, requirehittime, OD300, OD100, OD50, ODMiss)
	if isHit300(realhittime, requirehittime, OD300) {
		return Hit300
	}else if isHit100(realhittime, requirehittime, OD100) {
		return Hit100
	}else if isHit50(realhittime, requirehittime, OD50) {
		return Hit50
	}else if isHitMiss(realhittime, requirehittime, ODMiss) {
		return HitMiss
	}else {
		return HitMiss
	}
}

func isHitMiss(realhittime int64, requirehittime int64, ODMiss float64) bool {
	return (float64(realhittime) >= float64(requirehittime) - ODMiss) && (float64(realhittime) <= float64(requirehittime) + ODMiss)
}

func isHit50(realhittime int64, requirehittime int64, OD50 float64) bool {
	return (float64(realhittime) >= float64(requirehittime) - OD50) && (float64(realhittime) <= float64(requirehittime) + OD50)
}

func isHit100(realhittime int64, requirehittime int64, OD100 float64) bool {
	return (float64(realhittime) >= float64(requirehittime) - OD100) && (float64(realhittime) <= float64(requirehittime) + OD100)
}

func isHit300(realhittime int64, requirehittime int64, OD300 float64) bool {
	return (float64(realhittime) >= float64(requirehittime) - OD300) && (float64(realhittime) <= float64(requirehittime) + OD300)
}


