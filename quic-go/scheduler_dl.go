package quic

import (
	// "bitbucket.com/marcmolla/gorl"
	// "bitbucket.com/marcmolla/gorl/agents"
	// "bitbucket.com/marcmolla/gorl/types"
	// "errors"
	// "fmt"
	"github.com/lucas-clemente/quic-go/internal/protocol"
	// "github.com/lucas-clemente/quic-go/internal/utils"
	// "io/ioutil"
	"time"
	"math"
	// "github.com/nguyenthanhtrungbkhn/go-fuzzy-logic"
	"github.com/lucas-clemente/quic-go/internal/multiclients"

	// "os"
)
func roundFloat(val float64, precision uint) float64 {
    ratio := math.Pow(10, float64(precision))
    return math.Round(val*ratio) / ratio
}
// func GetAgent(weightsFile string, specFile string) agents.Agent {
// 	var spec []byte
// 	var err error
// 	if specFile != "" {
// 		spec, err = ioutil.ReadFile(specFile)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// 	agent := gorl.GetNormalInstance(string(spec))
// 	if weightsFile != "" {
// 		err = agent.LoadWeights(weightsFile)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// 	return agent
// }

// func GetTrainingAgent(weightsFile string, specFile string, outputPath string, epsilon float64) agents.TrainingAgent {
// 	var spec []byte
// 	var err error
// 	if specFile != "" {
// 		spec, err = ioutil.ReadFile(specFile)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}

// 	agent := gorl.GetTrainingInstance(string(spec), outputPath, float32(epsilon))
// 	if weightsFile != "" {
// 		err = agent.LoadWeights(weightsFile)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// 	return agent
// }

func NormalizeTimes(stat time.Duration) float32 {
	return float32(stat.Nanoseconds()) / float32(time.Millisecond.Nanoseconds())
}

func NormalizeGoodput(s *session, packetNumber uint64, retransNumber uint64) float32 {
	duration := time.Since(s.sessionCreationTime)

	// sentPackets := float32(packetNumber) * float32(protocol.DefaultTCPMSS)
	// retransPackets := float32(retransNumber) * float32(protocol.DefaultTCPMSS)

	elapsedtime := NormalizeTimes(duration) / 1000
	goodput := ((float32(packetNumber) - float32(retransNumber)) / 1024 / 1024 / elapsedtime) * float32(protocol.DefaultTCPMSS)
	
//	fmt.Println(packetNumber, retransNumber, elapsedtime, goodput)

	return goodput
}

// func RewardFinalGoodput(sch *scheduler, s *session, duration time.Duration, _ time.Duration) float32 {
// 	packetNumber := make(map[protocol.PathID]uint64)
// 	retransNumber := make(map[protocol.PathID]uint64)
// 	firstPath, secondPath := protocol.PathID(255), protocol.PathID(255)

// 	for pathID, path := range s.paths {
// 		if pathID != protocol.InitialPathID {
// 			packetNumber[pathID], retransNumber[pathID], _ = path.sentPacketHandler.GetStatistics()
// 			// Ordering paths
// 			if firstPath == protocol.PathID(255) {
// 				firstPath = pathID
// 			} else {
// 				if pathID < firstPath {
// 					secondPath = firstPath
// 					firstPath = pathID
// 				} else {
// 					secondPath = pathID
// 				}
// 			}
// 		}
// 	}

// 	sentPackets := float32(packetNumber[firstPath]+packetNumber[secondPath]) * float32(protocol.DefaultTCPMSS)
// 	retransPackets := float32(retransNumber[firstPath]+retransNumber[secondPath]) * float32(protocol.DefaultTCPMSS)

// 	elapsedtime := float32(duration)
// 	partialReward := (sentPackets - retransPackets) / 1024 / 1024 / elapsedtime
// 	//partialReward = float32(-100)

// 	return partialReward
// }

// func GetStateAndReward(sch *scheduler, s *session) (int, []*path) {
// 	packetNumber := make(map[protocol.PathID]uint64)
// 	retransNumber := make(map[protocol.PathID]uint64)

// 	sRTT := make(map[protocol.PathID]time.Duration)
// 	cwnd := make(map[protocol.PathID]protocol.ByteCount)
// 	cwndlevel := make(map[protocol.PathID]float32)

// 	firstPath, secondPath := protocol.PathID(255), protocol.PathID(255)

// 	for pathID, path := range s.paths {
// 		if pathID != protocol.InitialPathID {
// 			packetNumber[pathID], retransNumber[pathID], _ = path.sentPacketHandler.GetStatistics()
// 			sRTT[pathID] = path.rttStats.SmoothedRTT()
// 			cwnd[pathID] = path.sentPacketHandler.GetCongestionWindow()
// 			cwndlevel[pathID] = float32(path.sentPacketHandler.GetBytesInFlight()) / float32(cwnd[pathID])

// 			// Ordering paths
// 			if firstPath == protocol.PathID(255) {
// 				firstPath = pathID
// 			} else {
// 				if pathID < firstPath {
// 					secondPath = firstPath
// 					firstPath = pathID
// 				} else {
// 					secondPath = pathID
// 				}
// 			}
// 		}
// 	}

// 	//packetNumberInitial, _, _ := s.paths[protocol.InitialPathID].sentPacketHandler.GetStatistics()

// 	//Penalize and fast-quit
// 	// if sch.Training{
// 	// 	if packetNumberInitial > 20 {
// 	// 		utils.Errorf("closing: zero tolerance")
// 	// 		sch.TrainingAgent.CloseEpisode(uint64(s.connectionID), -100, false)
// 	// 		s.closeLocal(errors.New("closing: zero tolerance"))
// 	// 	}
// 	// }

// 	//State
// 	BSend, _ := s.flowControlManager.SendWindowSize(protocol.StreamID(5))
// 	state := types.Vector{NormalizeTimes(sRTT[firstPath]), NormalizeTimes(sRTT[secondPath]),
// 		float32(cwnd[firstPath]) / float32(protocol.DefaultTCPMSS) / 300, float32(cwnd[secondPath]) / float32(protocol.DefaultTCPMSS) / 300, cwndlevel[firstPath], cwndlevel[secondPath], float32(BSend) / float32(protocol.DefaultTCPMSS) / 300}

// 	//Action
// 	var action int
// 	// if sch.Training {
// 	// 	action = sch.TrainingAgent.GetAction(state)
// 	// } else {
// 	// 	action = sch.Agent.GetAction(state)
// 	// }

// 	//Write in state and action
// 	sch.statevector[sch.record] = state
// 	sch.actionvector[sch.record] = action

// 	//Partial Reward
// 	sentPackets := packetNumber[firstPath] + packetNumber[secondPath]
// 	retransPackets := retransNumber[firstPath] + retransNumber[secondPath]
// 	sch.packetvector[sch.record] = sentPackets - retransPackets

// 	partialReward := float32(0)
// 	elapsedtime := float32(0)
// 	buffertime := float32(0)
// 	sch.recordDuration[sch.record] = elapsedtime

// 	if sch.record == 0 {
// 		partialReward = float32(0)
// 		sch.episoderecord += 1
// 		if sch.Training {
// 			realstate := sch.statevector[sch.record]
// 			realaction := sch.actionvector[sch.record]
// 			sch.TrainingAgent.SaveStep(uint64(s.connectionID), partialReward, realstate, realaction)
// 		} else {
// 			if sch.DumpExp {
// 				sch.dumpAgent.AddStep(uint64(s.connectionID), []string{fmt.Sprint(sch.statevector[sch.record]), fmt.Sprint(sch.actionvector[sch.record])})
// 			}
// 		}
// 	} else {
// 		elapsedtime = float32(time.Since(sch.lastfiretime))
// 		sch.recordDuration[sch.record] = elapsedtime
// 		benchmark := sch.packetvector[sch.episoderecord-1]
// 		if benchmark < (sentPackets - retransPackets) {
// 			for i := uint64(0); i < (sentPackets - retransPackets - benchmark); i += 1 {
// 				for z := uint64(0); z < (sch.record - (sch.episoderecord - 1)); z += 1 {
// 					buffertime += sch.recordDuration[sch.episoderecord+z]
// 				}
// 				if sch.episoderecord == sch.record {
// 					partialReward = float32(sentPackets-retransPackets-benchmark-i) * float32(protocol.DefaultTCPMSS) / 1024 / 1024 / buffertime
// 					buffertime = float32(0)
// 					if sch.Training {
// 						realstate := sch.statevector[sch.episoderecord]
// 						realaction := sch.actionvector[sch.episoderecord]
// 						sch.TrainingAgent.SaveStep(uint64(s.connectionID), partialReward, realstate, realaction)
// 					} else {
// 						if sch.DumpExp {
// 							sch.dumpAgent.AddStep(uint64(s.connectionID), []string{fmt.Sprint(sch.statevector[sch.episoderecord]), fmt.Sprint(sch.actionvector[sch.episoderecord])})
// 						}
// 					}
// 					sch.episoderecord += 1
// 					break
// 				} else {
// 					partialReward = float32(protocol.DefaultTCPMSS) / 1024 / 1024 / buffertime
// 					buffertime = float32(0)
// 					if sch.Training {
// 						realstate := sch.statevector[sch.episoderecord]
// 						realaction := sch.actionvector[sch.episoderecord]
// 						sch.TrainingAgent.SaveStep(uint64(s.connectionID), partialReward, realstate, realaction)
// 					} else {
// 						if sch.DumpExp {
// 							sch.dumpAgent.AddStep(uint64(s.connectionID), []string{fmt.Sprint(sch.statevector[sch.episoderecord]), fmt.Sprint(sch.actionvector[sch.episoderecord])})
// 						}
// 					}
// 					sch.episoderecord += 1
// 				}
// 			}
// 		}
// 	}

// 	//Main pointer and fire time
// 	sch.record += 1
// 	sch.lastfiretime = time.Now()

// 	return action, []*path{s.paths[firstPath], s.paths[secondPath]}
// }

// func CheckAction(action int, state types.Vector, s *session, sch *scheduler) {
// 	if action != 0 {
// 		return
// 	}
// 	if state[4] < 1 || state[5] < 1 {
// 		// penalize not sending with one path allowed
// 		utils.Errorf("not sending with one path allowed")
// 		sch.TrainingAgent.CloseEpisode(uint64(s.connectionID), -100, false)
// 		s.closeLocal(errors.New("not sending with one path allowed"))
// 	}

// }

func (sch *scheduler) GetStateAndRewardQlearning(s *session, pth *path) {
	//rcvdpacketNumber := pth.lastRcvdPacketNumber
	packetNumber := make(map[protocol.PathID]uint64)
	retransNumber := make(map[protocol.PathID]uint64)

	lRTT := make(map[protocol.PathID]time.Duration)
	sRTT := make(map[protocol.PathID]time.Duration)

	cwnd := make(map[protocol.PathID]protocol.ByteCount)
	cwndlevel := make(map[protocol.PathID]float32)
	inp := make(map[protocol.PathID]protocol.ByteCount)

	reWard := make(map[protocol.PathID]float64)

	firstPath, secondPath := protocol.PathID(255), protocol.PathID(255)

	for pathID, path := range s.paths {
		//dataString := fmt.Sprintf("%d -", int(pathID))
		//f.WriteString(dataString)
		if pathID != protocol.InitialPathID {
			packetNumber[pathID], retransNumber[pathID], _ = path.sentPacketHandler.GetStatistics()
			lRTT[pathID] = path.rttStats.LatestRTT()
			sRTT[pathID] = path.rttStats.SmoothedRTT()
			cwnd[pathID] = path.sentPacketHandler.GetCongestionWindow()
			inp[pathID] = path.sentPacketHandler.GetBytesInFlight()
			if float32(cwnd[pathID]) != 0 {
				cwndlevel[pathID] = float32(path.sentPacketHandler.GetBytesInFlight()) / float32(cwnd[pathID])
			}else {
				cwndlevel[pathID] = 1
			}

			// Ordering paths
			if firstPath == protocol.PathID(255) {
				firstPath = pathID
			} else {
				if pathID < firstPath {
					secondPath = firstPath
					firstPath = pathID
				} else {
					secondPath = pathID
				}
			}
		}
	}
	if (s.scheduler.preRTT[pth.pathID] != 0) {
		s.scheduler.iRTT[pth.pathID] = 0.5 * s.scheduler.iRTT[pth.pathID] + 0.5 * float64(NormalizeTimes(lRTT[pth.pathID] - s.scheduler.preRTT[pth.pathID]))
	}
	s.scheduler.preRTT[pth.pathID] = lRTT[pth.pathID]

	//fmt.Println(s.scheduler.iRTT[pth.pathID])

	// goodput1 := NormalizeGoodput(s, packetNumber[firstPath], retransNumber[firstPath])
	// goodput2 := NormalizeGoodput(s, packetNumber[secondPath], retransNumber[secondPath])


	//fmt.Println(NormalizeTimes(lRTT[firstPath]), NormalizeTimes(maxRTT[firstPath]), NormalizeTimes(lRTT[secondPath]), NormalizeTimes(maxRTT[secondPath]))


	// f, err := os.OpenFile("/App/output/factor.data", os.O_APPEND|os.O_WRONLY, 0644)
	// if err != nil {
	// 	panic(err)
	// }
	// defer f.Close()
	// dataString6 := fmt.Sprintf("%d - %d - %d\n",pth.pathID, firstPath, secondPath)
	// dataString8 := fmt.Sprintf("%f - %f\n",s.scheduler.Alpha, s.scheduler.Beta)
	// dataString := fmt.Sprintf("%f - %f - %f\n",reWard[pth.pathID], reWard[firstPath], reWard[secondPath] )
	// dataString3 := fmt.Sprintf("%f - %f - %f\n---\n", float64(cwndlevel[pth.pathID]) , float64(NormalizeTimes(lRTT[pth.pathID])), float64(5*float32(retransNumber[pth.pathID])/float32(packetNumber[pth.pathID])))
	// f.WriteString(dataString6)
	// f.WriteString(dataString8)
	// f.WriteString(dataString)
	// f.WriteString(dataString3)

	//fmt.Println(s.scheduler.AdaDivisor, s.scheduler.Alpha, s.scheduler.Beta)
	if float32(packetNumber[firstPath]) > 0{
		reWard[firstPath] = (1-s.scheduler.Alpha - s.scheduler.Beta)*float64(cwndlevel[firstPath]) - s.scheduler.Alpha*float64(NormalizeTimes(lRTT[firstPath]) / 50) - s.scheduler.Beta*float64(5*float32(retransNumber[firstPath])/float32(packetNumber[firstPath]))
	}else{
		reWard[firstPath] = (1-s.scheduler.Alpha - s.scheduler.Beta)*float64(cwndlevel[firstPath]) - s.scheduler.Alpha*float64(NormalizeTimes(lRTT[firstPath]) / 50)
	}
	if float32(packetNumber[secondPath]) > 0{
		reWard[secondPath] = (1-s.scheduler.Alpha - s.scheduler.Beta)*float64(cwndlevel[secondPath]) - s.scheduler.Alpha*float64(NormalizeTimes(lRTT[secondPath]) / 50) - s.scheduler.Beta*float64(5*float32(retransNumber[secondPath])/float32(packetNumber[secondPath]))
	}else{
		reWard[secondPath] = (1-s.scheduler.Alpha - s.scheduler.Beta)*float64(cwndlevel[secondPath]) - s.scheduler.Alpha*float64(NormalizeTimes(lRTT[secondPath]) / 50)
	}
	

	//sendingRate := (float64(cwnd[pth.pathID])/ float64(lRTT[pth.pathID])) / (float64(cwnd[firstPath])/ float64(lRTT[firstPath]) + float64(cwnd[secondPath])/ float64(lRTT[secondPath]))
	f_sendingRate := (float64(cwnd[firstPath])/ float64(lRTT[firstPath])) / (float64(cwnd[firstPath])/ float64(lRTT[firstPath]) + float64(cwnd[secondPath])/ float64(lRTT[secondPath]))
	s_sendingRate := (float64(cwnd[secondPath])/ float64(lRTT[secondPath])) / (float64(cwnd[firstPath])/ float64(lRTT[firstPath]) + float64(cwnd[secondPath])/ float64(lRTT[secondPath]))


	//update Q
	var f_cLevel, s_cLevel, col int8

	if pth.pathID == firstPath {
		col = 0
		if reWard[firstPath] == 0 {
			return
		}
	}else {
		if reWard[secondPath] == 0 {
			return
		}
		col = 1
	}

	if f_sendingRate < sch.clv_state[0] {
		f_cLevel = 0
	}else if f_sendingRate < sch.clv_state[1] {
		f_cLevel = 1
	}else if f_sendingRate < sch.clv_state[2] {
		f_cLevel = 2
	}else if f_sendingRate < sch.clv_state[3] {
		f_cLevel = 3
	}else {
		f_cLevel = 4
	}

	if s_sendingRate < sch.clv_state[0] {
		s_cLevel = 0
	}else if s_sendingRate < sch.clv_state[1] {
		s_cLevel = 1
	}else if s_sendingRate < sch.clv_state[2] {
		s_cLevel = 2
	}else if s_sendingRate < sch.clv_state[3] {
		s_cLevel = 3
	}else {
		s_cLevel = 4
	}

	old_f_cLevel := s.scheduler.currentState_f
	old_s_cLevel := s.scheduler.currentState_s

	var maxNextState float64
	if s.scheduler.qtable[f_cLevel][s_cLevel][0] > s.scheduler.qtable[f_cLevel][s_cLevel][1] {
		maxNextState = s.scheduler.qtable[f_cLevel][s_cLevel][0]
	}else {
		maxNextState = s.scheduler.qtable[f_cLevel][s_cLevel][1]
	}
	// BSend, _ := s.flowControlManager.SendWindowSize(protocol.StreamID(5)) 

	//Fuzzy logic
	// if s.scheduler.SchedulerName == "fuzzyqsat"{
	// 	var data fuzzy.FuzzyNumber
	// 	data.Family.Number = string(s.scheduler.record)
	// 	data.Family.Income = float64(float32(cwnd[pth.pathID]))
	// 	data.Family.Debt = float64(float32(BSend))

	// 	blt := fuzzy.BLT{}
	// 	blt.Fuzzification(&data)
	// 	blt.Inference(&data)
	// 	blt.Defuzzification(&data)
	// 	if (s.scheduler.AdaDivisor != 1.0){
	// 		s.scheduler.Delta = 1 - data.CrispValue
	// 	} else{
	// 		s.scheduler.Delta = data.CrispValue
	// 	}
	// 	//fmt.Println(s.scheduler.Delta)
	// }
	s.scheduler.record += 1

	// fmt.Println("Vl: ", s.scheduler.Delta)	

	newValue := (1 - s.scheduler.Delta)*s.scheduler.qtable[old_f_cLevel][old_s_cLevel][col] + (s.scheduler.Delta)*(reWard[pth.pathID] + s.scheduler.Gamma*maxNextState)

	s.scheduler.qtable[old_f_cLevel][old_s_cLevel][col] = newValue
	s.scheduler.currentState_f = f_cLevel
	s.scheduler.currentState_s = s_cLevel
}

func (sch *scheduler) GetStateAndRewardMultiClients(s *session, pth *path) {
	packetNumber := make(map[protocol.PathID]uint64)
	retransNumber := make(map[protocol.PathID]uint64)

	lRTT := make(map[protocol.PathID]time.Duration)
	sRTT := make(map[protocol.PathID]time.Duration)

	cwnd := make(map[protocol.PathID]protocol.ByteCount)
	cwndlevel := make(map[protocol.PathID]float32)
	inp := make(map[protocol.PathID]protocol.ByteCount)

	reWard := make(map[protocol.PathID]float64)

	firstPath, secondPath := protocol.PathID(255), protocol.PathID(255)

	for pathID, path := range s.paths {
		if pathID != protocol.InitialPathID {
			packetNumber[pathID], retransNumber[pathID], _ = path.sentPacketHandler.GetStatistics()
			lRTT[pathID] = path.rttStats.LatestRTT()
			sRTT[pathID] = path.rttStats.SmoothedRTT()
			cwnd[pathID] = path.sentPacketHandler.GetCongestionWindow()
			inp[pathID] = path.sentPacketHandler.GetBytesInFlight()
			if float32(cwnd[pathID]) != 0 {
				cwndlevel[pathID] = float32(path.sentPacketHandler.GetBytesInFlight()) / float32(cwnd[pathID])
			}else {
				cwndlevel[pathID] = 1
			}

			// Ordering paths
			if firstPath == protocol.PathID(255) {
				firstPath = pathID
			} else {
				if pathID < firstPath {
					secondPath = firstPath
					firstPath = pathID
				} else {
					secondPath = pathID
				}
			}
		}
	}

	reWard[firstPath] = (1-s.scheduler.AdaDivisor)*float64(cwndlevel[firstPath]) - s.scheduler.Alpha * float64(NormalizeTimes(lRTT[firstPath]) )
	reWard[secondPath] = (1-s.scheduler.AdaDivisor)*float64(cwndlevel[secondPath]) - s.scheduler.Alpha * float64(NormalizeTimes(lRTT[secondPath]) )
	
	//Xac dinh trang thai cua mang
	f_sendingRate := (float64(cwnd[firstPath])/ float64(lRTT[firstPath])) / (float64(cwnd[firstPath])/ float64(lRTT[firstPath]) + float64(cwnd[secondPath])/ float64(lRTT[secondPath]))
	s_sendingRate := (float64(cwnd[secondPath])/ float64(lRTT[secondPath])) / (float64(cwnd[firstPath])/ float64(lRTT[firstPath]) + float64(cwnd[secondPath])/ float64(lRTT[secondPath]))
	
	fr_Rate := 0.0
	sr_Rate := 0.0

	if multiclients.S2.Count() > 1 {
		ItemsList := multiclients.S2.Items()
		for _, element := range ItemsList {
			if foo, ok := element.(multiclients.StateMulti); ok {
				fr_Rate += (float64(foo.FCWND) / float64(foo.FRTT)) / (float64(foo.FCWND) / float64(foo.FRTT) + float64(foo.SCWND) / float64(foo.SRTT))
				sr_Rate += (float64(foo.SCWND) / float64(foo.SRTT)) / (float64(foo.FCWND) / float64(foo.FRTT) + float64(foo.SCWND) / float64(foo.SRTT))
			}
	}
		fr_Rate = (fr_Rate - f_sendingRate) / float64(multiclients.S2.Count() - 1)
		sr_Rate = (sr_Rate - s_sendingRate) / float64(multiclients.S2.Count() - 1)
	}

	// tmp_para := 0.3
	// f_sendingRate = (1-tmp_para)*f_sendingRate + tmp_para*fr_Rate
	// s_sendingRate = (1-tmp_para)*s_sendingRate + tmp_para*sr_Rate
	var nf_cLevel, ns_cLevel, nfr_cLevel, nsr_cLevel int8

	if f_sendingRate < sch.clv_state[0] {
		nf_cLevel = 0
	}else if f_sendingRate < sch.clv_state[1] {
		nf_cLevel = 1
	}else if f_sendingRate < sch.clv_state[2] {
		nf_cLevel = 2
	}else if f_sendingRate < sch.clv_state[3] {
		nf_cLevel = 3
	}else {
		nf_cLevel = 4
	}

	if s_sendingRate < sch.clv_state[0] {
		ns_cLevel = 0
	}else if s_sendingRate < sch.clv_state[1] {
		ns_cLevel = 1
	}else if s_sendingRate < sch.clv_state[2] {
		ns_cLevel = 2
	}else if s_sendingRate < sch.clv_state[3] {
		ns_cLevel = 3
	}else {
		ns_cLevel = 4
	}

	if fr_Rate < sch.clv_state2[0] {
		nfr_cLevel = 0
	}else if fr_Rate < sch.clv_state2[1] {
		nfr_cLevel = 1
	}else if fr_Rate < sch.clv_state2[2] {
		nfr_cLevel = 2
	}else if fr_Rate < sch.clv_state2[3] {
		nfr_cLevel = 3
	}else {
		nfr_cLevel = 4
	}

	if sr_Rate >= sch.clv_state2[0] {
		nsr_cLevel = 0
	}else if sr_Rate >= sch.clv_state2[1] {
		nsr_cLevel = 1
	}else if sr_Rate >= sch.clv_state2[2] {
		nsr_cLevel = 2
	}else if sr_Rate >= sch.clv_state2[3] {
		nsr_cLevel = 3
	}else {
		nsr_cLevel = 4
	}


	//update Q follow by state of action t
	//Trang thai cu
	var f_cLevel, s_cLevel, fr_cLevel, sr_cLevel, col int8
	f_cLevel = sch.list_State[State{pth.pathID, pth.lastRcvdPacketNumber}].cState_f
	s_cLevel = sch.list_State[State{pth.pathID, pth.lastRcvdPacketNumber}].cState_s
	fr_cLevel = sch.list_State[State{pth.pathID, pth.lastRcvdPacketNumber}].cState_fr
	sr_cLevel = sch.list_State[State{pth.pathID, pth.lastRcvdPacketNumber}].cState_sr
	// f_cLevel = sch.currentState_f
	// s_cLevel = sch.currentState_s
	// fr_cLevel = sch.currentState_fr
	// sr_cLevel = sch.currentState_sr
	if pth.pathID == firstPath {
		col = 0
		if reWard[firstPath] == 0 {
			return
		}
	}else {
		if reWard[secondPath] == 0 {
			return
		}
		col = 1
	}

	// BSend, _ := s.flowControlManager.SendWindowSize(protocol.StreamID(5)) 

	//Fuzzy logic
	// if s.scheduler.SchedulerName == "multiclients"{
	// 	var data fuzzy.FuzzyNumber
	// 	data.Family.Number = string(s.scheduler.record)
	// 	data.Family.Income = float64(float32(cwnd[pth.pathID]))
	// 	data.Family.Debt = float64(float32(BSend))

	// 	blt := fuzzy.BLT{}
	// 	blt.Fuzzification(&data)
	// 	blt.Inference(&data)
	// 	blt.Defuzzification(&data)
	// 	if (s.scheduler.AdaDivisor != 1.0){
	// 		s.scheduler.Delta = 1 - data.CrispValue
	// 	} else{
	// 		s.scheduler.Delta = data.CrispValue
	// 	}
	// 	//fmt.Println(s.scheduler.Delta)
	// }
	s.scheduler.record += 1

	var maxNextState float64
	if multiclients.MultiQtable[nf_cLevel][ns_cLevel][nfr_cLevel][nsr_cLevel][0] > multiclients.MultiQtable[nf_cLevel][ns_cLevel][nfr_cLevel][nsr_cLevel][1] {
		maxNextState = multiclients.MultiQtable[nf_cLevel][ns_cLevel][nfr_cLevel][nsr_cLevel][0]
	}else {
		maxNextState = multiclients.MultiQtable[nf_cLevel][ns_cLevel][nfr_cLevel][nsr_cLevel][1]
	}
	
	newValue := (1 - s.scheduler.Delta)*multiclients.MultiQtable[s.scheduler.currentState_f][s.scheduler.currentState_s][s.scheduler.currentState_fr][s.scheduler.currentState_sr][col] + (s.scheduler.Delta)*(reWard[pth.pathID] + s.scheduler.Gamma*maxNextState)

	multiclients.MultiQtable[f_cLevel][s_cLevel][fr_cLevel][sr_cLevel][col] = newValue

	//fmt.Println("RewardAck: ", reWard[pth.pathID], maxNextState, newValue )
	// s.scheduler.currentState_f = f_cLevel
	// s.scheduler.currentState_s = s_cLevel
	// s.scheduler.currentState_fr = fr_cLevel
	// s.scheduler.currentState_sr = sr_cLevel
}

func (sch *scheduler) GetStateAndRewardMultiClientsRetrans(s *session, pth *path) {

	cwnd := make(map[protocol.PathID]protocol.ByteCount)
	firstPath, secondPath := protocol.PathID(255), protocol.PathID(255)

	for pathID, path := range s.paths {
		if pathID != protocol.InitialPathID {
			cwnd[pathID] = path.sentPacketHandler.GetCongestionWindow()
			// Ordering paths
			if firstPath == protocol.PathID(255) {
				firstPath = pathID
			} else {
				if pathID < firstPath {
					secondPath = firstPath
					firstPath = pathID
				} else {
					secondPath = pathID
				}
			}
		}
	}

	// BSend, _ := s.flowControlManager.SendWindowSize(protocol.StreamID(5)) 

	//Fuzzy logic
	// if s.scheduler.SchedulerName == "multiclients"{
	// 	var data fuzzy.FuzzyNumber
	// 	data.Family.Number = string(s.scheduler.record)
	// 	data.Family.Income = float64(float32(cwnd[pth.pathID]))
	// 	data.Family.Debt = float64(float32(BSend))

	// 	blt := fuzzy.BLT{}
	// 	blt.Fuzzification(&data)
	// 	blt.Inference(&data)
	// 	blt.Defuzzification(&data)
	// 	if (s.scheduler.AdaDivisor != 1.0){
	// 		s.scheduler.Delta = 1 - data.CrispValue
	// 	} else{
	// 		s.scheduler.Delta = data.CrispValue
	// 	}
	// 	//fmt.Println(s.scheduler.Delta)
	// }
	s.scheduler.record += 1


	reWard := make(map[protocol.PathID]float64)
	reWard[firstPath] = -s.scheduler.Beta
	reWard[secondPath] = -s.scheduler.Beta

	var f_cLevel, s_cLevel, fr_cLevel, sr_cLevel, col int8

	//update Q follow by state of action t

	f_cLevel = sch.list_State[State{pth.pathID, pth.lastRcvdPacketNumber}].cState_f
	s_cLevel = sch.list_State[State{pth.pathID, pth.lastRcvdPacketNumber}].cState_s
	fr_cLevel = sch.list_State[State{pth.pathID, pth.lastRcvdPacketNumber}].cState_fr
	sr_cLevel = sch.list_State[State{pth.pathID, pth.lastRcvdPacketNumber}].cState_sr

	if pth.pathID == firstPath {
		col = 0
	}else {
		col = 1
	}

	var maxNextState float64
	if multiclients.MultiQtable[f_cLevel][s_cLevel][fr_cLevel][sr_cLevel][0] > multiclients.MultiQtable[f_cLevel][s_cLevel][fr_cLevel][sr_cLevel][1] {
		maxNextState = multiclients.MultiQtable[f_cLevel][s_cLevel][fr_cLevel][sr_cLevel][0]
	}else {
		maxNextState = multiclients.MultiQtable[f_cLevel][s_cLevel][fr_cLevel][sr_cLevel][1]
	}
	
	newValue := (1 - s.scheduler.Delta)*multiclients.MultiQtable[s.scheduler.currentState_f][s.scheduler.currentState_s][s.scheduler.currentState_fr][s.scheduler.currentState_sr][col] + (s.scheduler.Delta)*(reWard[pth.pathID] + s.scheduler.Gamma*maxNextState)

	multiclients.MultiQtable[f_cLevel][s_cLevel][fr_cLevel][sr_cLevel][col] = newValue
	//fmt.Println("RewardRestran: ", reWard[pth.pathID], maxNextState, newValue )

}