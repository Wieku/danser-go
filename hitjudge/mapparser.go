package hitjudge

import (
	"danser/beatmap"
	"danser/beatmap/objects"
	"danser/bmath"
	"danser/replay"
	"danser/score"
	"github.com/Mempler/rplpa"
	"github.com/flesnuk/oppai5"
	"log"
	"math"
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

func ParseHits(mapname string, replayname string) ([]ObjectResult, []TotalResult) {
	// 加载map
	b := ParseMap(mapname)
	OD300 := b.OD300
	OD100 := b.OD100
	OD50 := b.OD50
	ODMiss := b.ODMiss
	convert_CS := 32 * (1 - 0.7 * (b.CircleSize - 5) / 5)
	// 加载replay
	pr := ParseReplay(replayname)
	r := pr.ReplayData

	// 如果replay是HR，改变OD和CS，并上下翻转replay的Y坐标
	if pr.Mods&16 > 0 {
		newOD := math.Min(1.4 * b.OD, 10)
		OD300 = beatmap.AdjustOD(79 - ( newOD * 6 ) + 0.5)
		OD100 = beatmap.AdjustOD(139 - ( newOD * 8 ) + 0.5)
		OD50 = beatmap.AdjustOD(199 - ( newOD * 10 ) + 0.5)
		ODMiss = beatmap.AdjustOD(229 - ( newOD * 11 ) + 0.5)
		convert_CS = 32 * (1 - 0.7 * (math.Min(1.3 * b.CircleSize, 10) - 5) / 5)
		makeReplayHR(r)
	}

	// 如果replay是EZ，改变OD和CS
	if pr.Mods&2 > 0 {
		newOD := b.OD / 2
		OD300 = beatmap.AdjustOD(79 - ( newOD * 6 ) + 0.5)
		OD100 = beatmap.AdjustOD(139 - ( newOD * 8 ) + 0.5)
		OD50 = beatmap.AdjustOD(199 - ( newOD * 10 ) + 0.5)
		ODMiss = beatmap.AdjustOD(229 - ( newOD * 11 ) + 0.5)
		convert_CS = 32 * (1 - 0.7 * (math.Min(b.CircleSize / 2, 10) - 5) / 5)
	}

	// 计数
	count300 := 0
	count100 := 0
	count50 := 0
	countMiss := 0

	// 结果数组
	result := []ObjectResult{}
	// 总体结果数组
	totalresult := []TotalResult{}
	// 判定数组
	totalhits := []int64{}
	// maxcombo
	maxcombo := 0
	nowcombo := 0

	// 依次处理HitObject
	keyindex := 3
	time := r[1].Time + r[2].Time
	for k := 0; k < len(b.HitObjects); k++ {
	//for k := 0; k < 2; k++ {
		//log.Println("Object", k+1)
		obj :=  b.HitObjects[k]
		if obj != nil {
			// 滑条
			if o, ok := obj.(*objects.Slider); ok {
				//log.Println("Slider info", o.GetBasicData().StartTime, o.GetBasicData().StartPos, o.GetBasicData().EndTime, o.GetBasicData().EndTime - o.TailJudgeOffset, o.TailJudgeOffset, o.TailJudgePoint, o.TickPoints)
				// 统计滑条的hit数，是否断连
				requirehits := 0
				realhits := 0
				isBreak := false
				// 判断滑条头
				requirehits += 1
				// ticks的判定倍数
				CS_scale := 2.4
				// 寻找最近的Key
				//log.Println("Slider head find", r[keyindex].Time, time, o.GetBasicData().StartTime, o.GetBasicData().StartPos)
				ticktime := 0.0
				if len(o.TickPoints) != 0 {
					ticktime = float64(o.TickPoints[0].Time)
				}else{
					ticktime = float64(o.GetBasicData().EndTime - o.TailJudgeOffset)
				}
				isfind, nearestindex, lasttime := findNearestKey(keyindex, time, r, o.GetBasicData().StartTime, o.GetBasicData().StartPos, ODMiss, OD50, convert_CS, true, ticktime)
				if isfind {
					// 如果找到，判断hit结果，设置下一个index+1
					keyhitresult := judgeHitResult(nearestindex, lasttime, r, o.GetBasicData().StartTime, ODMiss, OD300, OD100, OD50)
					switch keyhitresult {
					case Hit300:
						//log.Println("Slider head", o.GetBasicData().StartPos, o.GetBasicData().StartTime, "300")
						realhits += 1
						nowcombo += 1
						break
					case Hit100:
						//log.Println("Slider head", o.GetBasicData().StartPos, o.GetBasicData().StartTime, "100")
						realhits += 1
						nowcombo += 1
						break
					case Hit50:
						//log.Println("Slider head", o.GetBasicData().StartPos, o.GetBasicData().StartTime, "50")
						realhits += 1
						nowcombo += 1
						break
					case HitMiss:
						//log.Println("Slider head", o.GetBasicData().StartPos, o.GetBasicData().StartTime, "Miss")
						CS_scale = 1
						isBreak = true
						nowcombo = 0
						break
					}
					keyindex = nearestindex + 1
					time = lasttime + r[nearestindex].Time
					//log.Println("hit in", time)
				}else {
					// 如果没找到，输出miss，设置下一个index
					//log.Println("Slider head no found", o.GetBasicData().StartPos, o.GetBasicData().StartTime, "Miss", r[keyindex].Time, lasttime)
					CS_scale = 1
					isBreak = true
					nowcombo = 0
					keyindex = nearestindex
					time = lasttime
				}
				maxcombo = int(math.Max(float64(maxcombo), float64(nowcombo)))
				// 判断ticks
				for _, t := range o.ScorePoints {
					requirehits += 1
					isHit, nextindex, nexttime := isTickHit(keyindex, time, r, t.Time, t.Pos, CS_scale * convert_CS)
					keyindex = nextindex
					time = nexttime
					if isHit {
						//log.Println("Tick", i+1, "hit", t.Time, t.Pos)
						CS_scale = 2.4
						realhits += 1
						nowcombo += 1
					}else {
						//log.Println("Tick", i+1, "not hit", t.Time, t.Pos)
						CS_scale = 1
						isBreak = true
						nowcombo = 0
					}
					maxcombo = int(math.Max(float64(maxcombo), float64(nowcombo)))
				}
				// 判断滑条尾
				requirehits += 1
				//log.Println("Slider tail judge", r[keyindex - 1], time - r[keyindex - 1].Time, o.GetBasicData().EndTime - o.TailJudgeOffset, o.TailJudgeOffset, o.TailJudgePoint, CS_scale * convert_CS)
				isHit, nextindex, nexttime := isTickHit(keyindex - 1, time - r[keyindex - 1].Time, r, o.GetBasicData().EndTime - o.TailJudgeOffset, o.TailJudgePoint, CS_scale * convert_CS)

				if isHit {
					//log.Println("Slider tail hit", o.GetBasicData().EndTime, o.GetBasicData().EndPos)
					realhits += 1
					nowcombo += 1
					// 寻找状态改变后的时间点
					//log.Println("Start find slider release", r[nextindex].Time, nexttime+ r[nextindex].Time)
					keyindex, time = findRelease(nextindex, nexttime + r[nextindex].Time, r)
					time -= r[keyindex].Time
				}else {
					//log.Println("Slider tail not hit", o.GetBasicData().EndTime, o.GetBasicData().EndPos)
					//log.Println("Start find slider release", r[nextindex].Time, nexttime+ r[nextindex].Time)
					keyindex, time = findRelease(nextindex, nexttime + r[nextindex].Time, r)
					time -= r[keyindex].Time
				}
				maxcombo = int(math.Max(float64(maxcombo), float64(nowcombo)))
				// 滑条总体情况
				sliderhitresult := judgeSlider(requirehits, realhits)
				switch sliderhitresult {
				case Hit300:
					//log.Println("Slider count as 300", requirehits, realhits)
					count300 += 1
					totalhits = append(totalhits, 300)
					realhits += 1
					break
				case Hit100:
					//log.Println("Slider count as 100", requirehits, realhits)
					count100 += 1
					totalhits = append(totalhits, 100)
					realhits += 1
					break
				case Hit50:
					//log.Println("Slider count as 50", requirehits, realhits)
					count50 += 1
					totalhits = append(totalhits, 50)
					realhits += 1
					break
				case HitMiss:
					//log.Println("Slider count as Miss", requirehits, realhits)
					countMiss += 1
					totalhits = append(totalhits, 0)
					isBreak = true
					break
				}
				if isBreak {
					//log.Println("Slider breaks")
				}else {
					//log.Println("Slider no breaks")
				}
				result = append(result, ObjectResult{o.GetBasicData().StartPos, o.GetBasicData().EndTime - o.TailJudgeOffset, sliderhitresult, isBreak})
			}
			// note
			if o, ok := obj.(*objects.Circle); ok {
				// 寻找最近的Key
				keyhitresult := HitMiss
				isBreak := true
				isfind, nearestindex, lasttime := findNearestKey(keyindex, time, r, o.GetBasicData().StartTime, o.GetBasicData().StartPos, ODMiss, OD50, convert_CS, false, 0)
				if isfind {
					// 如果找到，判断hit结果，设置下一个index+1
					keyhitresult = judgeHitResult(nearestindex, lasttime, r, o.GetBasicData().StartTime, ODMiss, OD300, OD100, OD50)
					switch keyhitresult {
					case Hit300:
						//log.Println("Circle count as 300")
						count300 += 1
						nowcombo += 1
						totalhits = append(totalhits, 300)
						break
					case Hit100:
						//log.Println("Circle count as 100")
						count100 += 1
						nowcombo += 1
						totalhits = append(totalhits, 100)
						break
					case Hit50:
						//log.Println("Circle count as 50")
						count50 += 1
						nowcombo += 1
						totalhits = append(totalhits, 50)
						break
					case HitMiss:
						//log.Println("Circle count as Miss")
						countMiss += 1
						nowcombo = 0
						totalhits = append(totalhits, 0)
						break
					}
					time = lasttime + r[nearestindex].Time
					//log.Println("hit in", time)
					// 寻找状态改变后的时间点
					keyindex, time = findRelease(nearestindex, time, r)
					time -= r[keyindex].Time
				}else {
					// 如果没找到，输出miss，设置下一个index
					//log.Println("Circle count as Miss")
					countMiss += 1
					nowcombo = 0
					keyindex = nearestindex
					time = lasttime
				}
				if keyhitresult != HitMiss {
					isBreak = false
				}
				maxcombo = int(math.Max(float64(maxcombo), float64(nowcombo)))
				result = append(result, ObjectResult{o.GetBasicData().StartPos, o.GetBasicData().StartTime, keyhitresult, isBreak})
			}
			// 转盘
			if o, ok := obj.(*objects.Spinner); ok {
				//log.Println("Spinner! skip!", o.GetBasicData())
				count300 += 1
				nowcombo += 1
				totalhits = append(totalhits, 300)
				maxcombo = int(math.Max(float64(maxcombo), float64(nowcombo)))
				result = append(result, ObjectResult{o.GetBasicData().StartPos, o.GetBasicData().StartTime, Hit300, false})
			}
		}
		tmptotalresult := TotalResult{	uint16(count300),
										uint16(count100),
										uint16(count50),
										uint16(countMiss),
										uint16(maxcombo),
										pr.Mods,
										score.CalculateAccuracy(totalhits),
										score.CalculateRank(totalhits),
										oppai.PPv2{}}
		//tmptotalresult.PP = calculatePP(mapname, tmptotalresult)
		tmptotalresult.PP = calculatePPbyNum(mapname, tmptotalresult, k+1)
		totalresult = append(totalresult, tmptotalresult)
		//log.Println("Now Max Combo:", maxcombo)
		//log.Println("Acc:", score.CalculateAccuracy(totalhits))
	}

	log.Println("Count 300:", count300)
	log.Println("Count 100:", count100)
	log.Println("Count 50:", count50)
	log.Println("Count Miss:", countMiss)
	log.Println("Max Combo:", maxcombo)
	log.Println("Acc:", totalresult[len(totalresult)-1].Acc)

	return result, totalresult
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
		//log.Println("Key compare", time - r[index].Time, *keypress, time, *r[index].KeyPressed, isPressChanged(*keypress, *r[index].KeyPressed))
		//if time > 29400 {
		//	os.Exit(2)
		//}
		if isPressChanged(*keypress, *r[index].KeyPressed) {
			//log.Println("Find release before", r[index].Time, time)
			return index, time
		}
		keypress = r[index].KeyPressed
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
func findNearestKey(start int, starttime int64, r []*rplpa.ReplayData, requirehittime int64, requirepos bmath.Vector2d, ODMiss float64, OD50 float64, CS float64, isNextTick bool, ticktime float64) (bool, int, int64) {
	index := start
	time := starttime
	for {
		hit := r[index]
		//log.Println("Find move", hit.Time + time, requirehittime, isInCircle(hit, requirepos, CS), isPressed(hit), bmath.NewVec2d(float64(hit.MosueX), float64(hit.MouseY)), requirepos, bmath.Vector2d.Dst(bmath.NewVec2d(float64(hit.MosueX), float64(hit.MouseY)), requirepos), ODMiss, OD50, CS + 0.05)
		//if hit.Time + time > 8300 {
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
				}
			}
		}else {
			// 如果不在圈内，且按下按键
			if isPressed(hit) {
				realhittime := hit.Time + time
				// 判断这个时间点和object时间点的关系
				if float64(realhittime) > float64(requirehittime) + OD50 {
					// 如果在最后时间之后按下，没效果，等于没找到，返回这个Key的时间点
					// 最后时间为最后能按出50的时间
					//log.Println("Hit too late", realhittime, requirehittime)
					// （tick、滑条尾）返回上一个生效点
					if isNextTick {
						index, time = findFirstAfterLastHit(ticktime, r)
						time -= r[index].Time
						//log.Println("Return to last tick point", r[index].Time, time)
					}
					return false, index, time
				}else {
					// 如果最后时间前按下，没效果，此键位失去对下一个非tick的object（note、滑条头）的效果，寻找下一个按键放下的地方
					//log.Println("Tap out is no use!")
					index, time = findRelease(index, realhittime, r)
					time -= r[index].Time
					// （tick、滑条尾）如果这个时间大于最后时间，则用最后时间重新定位tick生效位置
					if float64(time) > float64(requirehittime) + OD50 {
						if isNextTick {
							index, time = findFirstAfterLastHit(ticktime, r)
							time -= r[index].Time
							//log.Println("Return to last tick point", r[index].Time, time)
						}
						return false, index, time
					}
					continue
				}
			}
		}
		index++
		time += hit.Time
	}
}

// 该时间点是否按下按键
func isPressed(hit *rplpa.ReplayData) bool {
	press := hit.KeyPressed
	return press.LeftClick || press.RightClick || press.Key1 || press.Key2
}

func isInCircle(hit *rplpa.ReplayData, requirepos bmath.Vector2d, CS float64) bool {
	realpos := bmath.NewVec2d(float64(hit.MosueX), float64(hit.MouseY))
	// 加入少量误差
	return bmath.Vector2d.Dst(realpos, requirepos) <= (CS + 0.05)
}

// 是否超过object的最后时间点
func isHitOver(realhittime int64, requirehittime int64, ODMiss float64) bool {
	return float64(realhittime) > float64(requirehittime) + ODMiss
}

// 判断hit结果
func judgeHitResult(index int, lasttime int64, r []*rplpa.ReplayData, requirehittime int64, ODMiss float64, OD300 float64, OD100 float64, OD50 float64) HitResult{
	realhittime := r[index].Time + lasttime
	//log.Println("Judge Hit", realhittime, requirehittime, OD300, OD100, OD50, ODMiss)
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

// 判断tick是否被击中并按下
func isTickHit(start int, starttime int64, r []*rplpa.ReplayData, requirehittime int64, requirepos bmath.Vector2d, CS float64) (bool, int, int64) {
	index := start - 1
	time := starttime - r[index].Time
	for {
		//寻找正好的一点或者区间
		//log.Println("Judge index", index)
		hit := r[index]
		realhittime := hit.Time + time
		if realhittime == requirehittime {
			// 找到正好的一点
			//log.Println("Tick Judge Tap", requirehittime, realhittime, bmath.NewVec2d(float64(hit.MosueX), float64(hit.MouseY)), requirepos, bmath.Vector2d.Dst(bmath.NewVec2d(float64(hit.MosueX), float64(hit.MouseY)), requirepos), CS)
			if isInCircle(hit, requirepos, CS) {
				// 在圈内
				if isPressed(hit) {
					//按下，则击中成功
					return true, index + 1, realhittime
				}
			}
			return false, index + 1, realhittime
		}else if realhittime < requirehittime && realhittime + r[index+1].Time > requirehittime{
			// 找到正好的区间
			//log.Println("Tick Judge Range", requirehittime, realhittime, realhittime + r[index+1].Time, bmath.NewVec2d(float64(hit.MosueX), float64(hit.MouseY)), requirepos, bmath.Vector2d.Dst(bmath.NewVec2d(float64(hit.MosueX), float64(hit.MouseY)), requirepos), CS)
			if isInCircle(hit, requirepos, CS) {
				// 前一点在圈内
				if isPressed(hit) {
					//前一点按下，则击中成功
					return true, index + 1, realhittime
				}
			}
			return false, index + 1, realhittime
		}else if realhittime > requirehittime {
			// 时间点已经超过需要的击中时间，则已经无法击中
			//log.Println("Too late to hit tick", realhittime, requirehittime)
			return false, index, realhittime - hit.Time
		}
		index++
		time += hit.Time
	}
}

// 判断滑条最终情况
func judgeSlider(requirehits int, realhits int) HitResult {
	// 一个滑条的击中比例
	hitfraction := float64(realhits) / float64(requirehits)
	if hitfraction==1 {
		// 击中比例等于1，输出300
		return Hit300
	}else if hitfraction >=0.5 {
		// 击中比例大于等于0.5，输出100
		return Hit100
	}else if hitfraction >0 {
		// 击中比例大于0，输出50
		return Hit50
	}else {
		// 击中比例为0，输出miss
		return HitMiss
	}
}

// 通过最后时间找第一个tick生效位置
func findFirstAfterLastHit(ticktime float64, r []*rplpa.ReplayData) (int, int64) {
	index := 3
	time := r[1].Time + r[2].Time
	for {
		time += r[index].Time
		if float64(time) > ticktime {
			//log.Println("Find FirstbeforeTick before", r[index].Time, time, ticktime)
			time -= r[index].Time
			return index - 1, time
		}
		index++
	}
}

// HR上下翻转replay
func makeReplayHR(r []*rplpa.ReplayData){
	for k := 0; k < len(r); k++ {
		r[k].MouseY = 384 - r[k].MouseY
	}
}

// oppai载入map
func loadMap(filename string) *oppai.Map {
	f, _ := os.Open(filename)
	return oppai.Parse(f)
}

// 部分载入map
func loadMapbyNum(filename string, objnum int) *oppai.Map {
	f, _ := os.Open(filename)
	return oppai.ParsebyNum(f, objnum)
}

// oppai计算pp
func calculatePP(filename string, result TotalResult) oppai.PPv2 {
	//amap := loadMap(filename)
	//diff := (&oppai.DiffCalc{}).CalcMapWithMods(*amap, int(result.Mods))
	//log.Println(diff.Beatmap.MaxCombo, result.Combo, diff.Total)
	return oppai.PPInfo(loadMap(filename), &oppai.Parameters{
		Combo:  result.Combo,
		Mods:   result.Mods,
		N300:   result.N300,
		N100:   result.N100,
		N50:    result.N50,
		Misses: result.Misses,
	}).PP
}

// 计算部分的pp
func calculatePPbyNum(filename string, result TotalResult, objnum int) oppai.PPv2 {
	//amap := loadMapbyNum(filename, objnum)
	//diff := (&oppai.DiffCalc{}).CalcMapWithMods(*amap, int(result.Mods))
	//log.Println(diff.Beatmap.MaxCombo, result.Combo, diff.Total)
	return oppai.PPInfo(loadMapbyNum(filename, objnum), &oppai.Parameters{
		Combo:  result.Combo,
		Mods:   result.Mods,
		N300:   result.N300,
		N100:   result.N100,
		N50:    result.N50,
		Misses: result.Misses,
	}).PP
}
