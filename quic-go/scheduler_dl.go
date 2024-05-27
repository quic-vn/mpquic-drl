package quic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/lucas-clemente/quic-go/internal/protocol"

	// "github.com/nguyenthanhtrungbkhn/go-fuzzy-logic"
	"github.com/lucas-clemente/quic-go/internal/multiclients"
)

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func NormalizeTimes(stat time.Duration) float32 {
	return float32(stat.Nanoseconds()) / float32(time.Millisecond.Nanoseconds())
}

func NormalizeGoodput(s *session, packetNumber uint64, retransNumber uint64) float32 {
	duration := time.Since(s.sessionCreationTime)

	elapsedtime := NormalizeTimes(duration) / 1000
	goodput := ((float32(packetNumber) - float32(retransNumber)) / 1024 / 1024 / elapsedtime) * float32(protocol.DefaultTCPMSS)

	return goodput
}

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
			} else {
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
	if s.scheduler.preRTT[pth.pathID] != 0 {
		s.scheduler.iRTT[pth.pathID] = 0.5*s.scheduler.iRTT[pth.pathID] + 0.5*float64(NormalizeTimes(lRTT[pth.pathID]-s.scheduler.preRTT[pth.pathID]))
	}
	s.scheduler.preRTT[pth.pathID] = lRTT[pth.pathID]

	if float32(packetNumber[firstPath]) > 0 {
		reWard[firstPath] = (1-s.scheduler.Alpha-s.scheduler.Beta)*float64(cwndlevel[firstPath]) - s.scheduler.Alpha*float64(NormalizeTimes(lRTT[firstPath])/50) - s.scheduler.Beta*float64(5*float32(retransNumber[firstPath])/float32(packetNumber[firstPath]))
	} else {
		reWard[firstPath] = (1-s.scheduler.Alpha-s.scheduler.Beta)*float64(cwndlevel[firstPath]) - s.scheduler.Alpha*float64(NormalizeTimes(lRTT[firstPath])/50)
	}
	if float32(packetNumber[secondPath]) > 0 {
		reWard[secondPath] = (1-s.scheduler.Alpha-s.scheduler.Beta)*float64(cwndlevel[secondPath]) - s.scheduler.Alpha*float64(NormalizeTimes(lRTT[secondPath])/50) - s.scheduler.Beta*float64(5*float32(retransNumber[secondPath])/float32(packetNumber[secondPath]))
	} else {
		reWard[secondPath] = (1-s.scheduler.Alpha-s.scheduler.Beta)*float64(cwndlevel[secondPath]) - s.scheduler.Alpha*float64(NormalizeTimes(lRTT[secondPath])/50)
	}

	//sendingRate := (float64(cwnd[pth.pathID])/ float64(lRTT[pth.pathID])) / (float64(cwnd[firstPath])/ float64(lRTT[firstPath]) + float64(cwnd[secondPath])/ float64(lRTT[secondPath]))
	f_sendingRate := (float64(cwnd[firstPath]) / float64(lRTT[firstPath])) / (float64(cwnd[firstPath])/float64(lRTT[firstPath]) + float64(cwnd[secondPath])/float64(lRTT[secondPath]))
	s_sendingRate := (float64(cwnd[secondPath]) / float64(lRTT[secondPath])) / (float64(cwnd[firstPath])/float64(lRTT[firstPath]) + float64(cwnd[secondPath])/float64(lRTT[secondPath]))

	//update Q
	var f_cLevel, s_cLevel, col int8

	if pth.pathID == firstPath {
		col = 0
		if reWard[firstPath] == 0 {
			return
		}
	} else {
		if reWard[secondPath] == 0 {
			return
		}
		col = 1
	}

	if f_sendingRate < sch.clv_state[0] {
		f_cLevel = 0
	} else if f_sendingRate < sch.clv_state[1] {
		f_cLevel = 1
	} else if f_sendingRate < sch.clv_state[2] {
		f_cLevel = 2
	} else if f_sendingRate < sch.clv_state[3] {
		f_cLevel = 3
	} else {
		f_cLevel = 4
	}

	if s_sendingRate < sch.clv_state[0] {
		s_cLevel = 0
	} else if s_sendingRate < sch.clv_state[1] {
		s_cLevel = 1
	} else if s_sendingRate < sch.clv_state[2] {
		s_cLevel = 2
	} else if s_sendingRate < sch.clv_state[3] {
		s_cLevel = 3
	} else {
		s_cLevel = 4
	}

	old_f_cLevel := s.scheduler.currentState_f
	old_s_cLevel := s.scheduler.currentState_s

	var maxNextState float64
	if s.scheduler.qtable[f_cLevel][s_cLevel][0] > s.scheduler.qtable[f_cLevel][s_cLevel][1] {
		maxNextState = s.scheduler.qtable[f_cLevel][s_cLevel][0]
	} else {
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

	newValue := (1-s.scheduler.Delta)*s.scheduler.qtable[old_f_cLevel][old_s_cLevel][col] + (s.scheduler.Delta)*(reWard[pth.pathID]+s.scheduler.Gamma*maxNextState)

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
			} else {
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

	reWard[firstPath] = (1-s.scheduler.AdaDivisor)*float64(cwndlevel[firstPath]) - s.scheduler.Alpha*float64(NormalizeTimes(lRTT[firstPath]))
	reWard[secondPath] = (1-s.scheduler.AdaDivisor)*float64(cwndlevel[secondPath]) - s.scheduler.Alpha*float64(NormalizeTimes(lRTT[secondPath]))

	//Xac dinh trang thai cua mang
	f_sendingRate := (float64(cwnd[firstPath]) / float64(lRTT[firstPath])) / (float64(cwnd[firstPath])/float64(lRTT[firstPath]) + float64(cwnd[secondPath])/float64(lRTT[secondPath]))
	s_sendingRate := (float64(cwnd[secondPath]) / float64(lRTT[secondPath])) / (float64(cwnd[firstPath])/float64(lRTT[firstPath]) + float64(cwnd[secondPath])/float64(lRTT[secondPath]))

	fr_Rate := 0.0
	sr_Rate := 0.0

	if multiclients.S2.Count() > 1 {
		ItemsList := multiclients.S2.Items()
		for _, element := range ItemsList {
			if foo, ok := element.(multiclients.StateMulti); ok {
				fr_Rate += (float64(foo.FCWND) / float64(foo.FRTT)) / (float64(foo.FCWND)/float64(foo.FRTT) + float64(foo.SCWND)/float64(foo.SRTT))
				sr_Rate += (float64(foo.SCWND) / float64(foo.SRTT)) / (float64(foo.FCWND)/float64(foo.FRTT) + float64(foo.SCWND)/float64(foo.SRTT))
			}
		}
		fr_Rate = (fr_Rate - f_sendingRate) / float64(multiclients.S2.Count()-1)
		sr_Rate = (sr_Rate - s_sendingRate) / float64(multiclients.S2.Count()-1)
	}

	// tmp_para := 0.3
	// f_sendingRate = (1-tmp_para)*f_sendingRate + tmp_para*fr_Rate
	// s_sendingRate = (1-tmp_para)*s_sendingRate + tmp_para*sr_Rate
	var nf_cLevel, ns_cLevel, nfr_cLevel, nsr_cLevel int8

	if f_sendingRate < sch.clv_state[0] {
		nf_cLevel = 0
	} else if f_sendingRate < sch.clv_state[1] {
		nf_cLevel = 1
	} else if f_sendingRate < sch.clv_state[2] {
		nf_cLevel = 2
	} else if f_sendingRate < sch.clv_state[3] {
		nf_cLevel = 3
	} else {
		nf_cLevel = 4
	}

	if s_sendingRate < sch.clv_state[0] {
		ns_cLevel = 0
	} else if s_sendingRate < sch.clv_state[1] {
		ns_cLevel = 1
	} else if s_sendingRate < sch.clv_state[2] {
		ns_cLevel = 2
	} else if s_sendingRate < sch.clv_state[3] {
		ns_cLevel = 3
	} else {
		ns_cLevel = 4
	}

	if fr_Rate < sch.clv_state2[0] {
		nfr_cLevel = 0
	} else if fr_Rate < sch.clv_state2[1] {
		nfr_cLevel = 1
	} else if fr_Rate < sch.clv_state2[2] {
		nfr_cLevel = 2
	} else if fr_Rate < sch.clv_state2[3] {
		nfr_cLevel = 3
	} else {
		nfr_cLevel = 4
	}

	if sr_Rate >= sch.clv_state2[0] {
		nsr_cLevel = 0
	} else if sr_Rate >= sch.clv_state2[1] {
		nsr_cLevel = 1
	} else if sr_Rate >= sch.clv_state2[2] {
		nsr_cLevel = 2
	} else if sr_Rate >= sch.clv_state2[3] {
		nsr_cLevel = 3
	} else {
		nsr_cLevel = 4
	}

	//update Q follow by state of action t
	//Trang thai cu
	var f_cLevel, s_cLevel, fr_cLevel, sr_cLevel, col int8
	f_cLevel = sch.list_State[State{pth.pathID, pth.lastRcvdPacketNumber, sch.current_Prob}].cState_f
	s_cLevel = sch.list_State[State{pth.pathID, pth.lastRcvdPacketNumber, sch.current_Prob}].cState_s
	fr_cLevel = sch.list_State[State{pth.pathID, pth.lastRcvdPacketNumber, sch.current_Prob}].cState_fr
	sr_cLevel = sch.list_State[State{pth.pathID, pth.lastRcvdPacketNumber, sch.current_Prob}].cState_sr
	// f_cLevel = sch.currentState_f
	// s_cLevel = sch.currentState_s
	// fr_cLevel = sch.currentState_fr
	// sr_cLevel = sch.currentState_sr
	if pth.pathID == firstPath {
		col = 0
		if reWard[firstPath] == 0 {
			return
		}
	} else {
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
	} else {
		maxNextState = multiclients.MultiQtable[nf_cLevel][ns_cLevel][nfr_cLevel][nsr_cLevel][1]
	}

	newValue := (1-s.scheduler.Delta)*multiclients.MultiQtable[s.scheduler.currentState_f][s.scheduler.currentState_s][s.scheduler.currentState_fr][s.scheduler.currentState_sr][col] + (s.scheduler.Delta)*(reWard[pth.pathID]+s.scheduler.Gamma*maxNextState)

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

	f_cLevel = sch.list_State[State{pth.pathID, pth.lastRcvdPacketNumber, sch.current_Prob}].cState_f
	s_cLevel = sch.list_State[State{pth.pathID, pth.lastRcvdPacketNumber, sch.current_Prob}].cState_s
	fr_cLevel = sch.list_State[State{pth.pathID, pth.lastRcvdPacketNumber, sch.current_Prob}].cState_fr
	sr_cLevel = sch.list_State[State{pth.pathID, pth.lastRcvdPacketNumber, sch.current_Prob}].cState_sr

	if pth.pathID == firstPath {
		col = 0
	} else {
		col = 1
	}

	var maxNextState float64
	if multiclients.MultiQtable[f_cLevel][s_cLevel][fr_cLevel][sr_cLevel][0] > multiclients.MultiQtable[f_cLevel][s_cLevel][fr_cLevel][sr_cLevel][1] {
		maxNextState = multiclients.MultiQtable[f_cLevel][s_cLevel][fr_cLevel][sr_cLevel][0]
	} else {
		maxNextState = multiclients.MultiQtable[f_cLevel][s_cLevel][fr_cLevel][sr_cLevel][1]
	}

	newValue := (1-s.scheduler.Delta)*multiclients.MultiQtable[s.scheduler.currentState_f][s.scheduler.currentState_s][s.scheduler.currentState_fr][s.scheduler.currentState_sr][col] + (s.scheduler.Delta)*(reWard[pth.pathID]+s.scheduler.Gamma*maxNextState)

	multiclients.MultiQtable[f_cLevel][s_cLevel][fr_cLevel][sr_cLevel][col] = newValue
	//fmt.Println("RewardRestran: ", reWard[pth.pathID], maxNextState, newValue )

}

// func updateReward(url string, payload RewardPayload) error {
// 	jsonPayload, err := json.Marshal(payload)
// 	if err != nil {
// 		return err
// 	}

// 	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	var response map[string]interface{}
// 	err = json.NewDecoder(resp.Body).Decode(&response)
// 	if err != nil {
// 		return err
// 	}

// 	if status, ok := response["status"].(string); !ok || status != "Reward updated" {
// 		return fmt.Errorf("unexpected response: %v", response)
// 	}

// 	return nil
// }

func updateReward(url string, payload RewardPayload) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Use a goroutine to perform the POST request without waiting for the response
	go func() {
		_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			fmt.Println("Error sending POST request:", err)
		}
	}()

	return nil
}

func (sch *scheduler) GetStateAndRewardDQN(s *session, pth *path) {
	packetNumber := make(map[protocol.PathID]uint64)
	retransNumber := make(map[protocol.PathID]uint64)
	lostNumber := make(map[protocol.PathID]uint64)

	lRTT := make(map[protocol.PathID]time.Duration)
	sRTT := make(map[protocol.PathID]time.Duration)
	mRTT := make(map[protocol.PathID]time.Duration)

	CWND := make(map[protocol.PathID]protocol.ByteCount)
	CWNDlevel := make(map[protocol.PathID]float32)
	INP := make(map[protocol.PathID]protocol.ByteCount)

	reWard := make(map[protocol.PathID]float64)

	firstPath, secondPath := protocol.PathID(255), protocol.PathID(255)

	for pathID, path := range s.paths {
		if pathID != protocol.InitialPathID {
			packetNumber[pathID], retransNumber[pathID], lostNumber[pathID] = path.sentPacketHandler.GetStatistics()
			lRTT[pathID] = path.rttStats.LatestRTT()
			sRTT[pathID] = path.rttStats.SmoothedRTT()
			mRTT[pathID] = path.rttStats.MinRTT()
			CWND[pathID] = path.sentPacketHandler.GetCongestionWindow()
			INP[pathID] = path.sentPacketHandler.GetBytesInFlight()
			if float32(CWND[pathID]) != 0 {
				CWNDlevel[pathID] = float32(path.sentPacketHandler.GetBytesInFlight()) / float32(CWND[pathID])
			} else {
				CWNDlevel[pathID] = 1
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

	//alpha := 0.01
	//reWard[firstPath] = goodput1 - alpha*float64(NormalizeTimes(lRTT[firstPath])) - float64(lostNumber[firstPath]/packetNumber[firstPath])
	//reWard[secondPath] = goodput1 - alpha*float64(NormalizeTimes(lRTT[secondPath]))- float64(lostNumber[secondPath]/packetNumber[secondPath])

	rttrate := 0.0
	goodput := NormalizeGoodput(s, packetNumber[pth.pathID], retransNumber[pth.pathID])
	lostrate := 10 * float64(lostNumber[pth.pathID]) / float64(packetNumber[pth.pathID])
	if mRTT[pth.pathID] != 0 {
		rttrate = float64(lRTT[pth.pathID]) / float64(mRTT[pth.pathID])
	}

	reWard[pth.pathID] = float64(goodput) - rttrate - lostrate
	// fmt.Println("reWard", float64(goodput), rttrate, lostrate)
	old_state := sch.list_State_DQN[State{pth.pathID, pth.lastRcvdPacketNumber, s.scheduler.current_Prob}]

	nextState := StateDQN{
		CWNDf: float64(CWND[firstPath]),
		INPf:  float64(INP[firstPath]),
		SRTTf: float64(sRTT[firstPath]),
		CWNDs: float64(CWND[secondPath]),
		INPs:  float64(INP[secondPath]),
		SRTTs: float64(sRTT[secondPath]),
	}

	rewardPayload := RewardPayload{
		State:     old_state,
		NextState: nextState,
		Action:    sch.current_Prob,
		Reward:    reWard[pth.pathID],
		Done:      false,
	}

	err := updateReward(baseURL+"/update_reward", rewardPayload)
	if err != nil {
		fmt.Println("Error updating reward:", err)
		return
	}
}

func (sch *scheduler) GetStateAndRewardQSAT(s *session, pth *path) {
	rcvdpacketNumber := pth.lastRcvdPacketNumber
	packetNumber := make(map[protocol.PathID]uint64)
	retransNumber := make(map[protocol.PathID]uint64)

	sRTT := make(map[protocol.PathID]time.Duration)
	maxRTT := make(map[protocol.PathID]time.Duration)

	cwnd := make(map[protocol.PathID]protocol.ByteCount)
	cwndlevel := make(map[protocol.PathID]float32)

	reWard := make(map[protocol.PathID]float64)

	firstPath, secondPath := protocol.PathID(255), protocol.PathID(255)

	for pathID, path := range s.paths {
		//dataString := fmt.Sprintf("%d -", int(pathID))
		//f.WriteString(dataString)
		if pathID != protocol.InitialPathID {
			packetNumber[pathID], retransNumber[pathID], _ = path.sentPacketHandler.GetStatistics()
			sRTT[pathID] = path.rttStats.LatestRTT()
			maxRTT[pathID] = path.rttStats.MaxRTT()
			cwnd[pathID] = path.sentPacketHandler.GetCongestionWindow()
			if float32(cwnd[pathID]) != 0 {
				cwndlevel[pathID] = float32(path.sentPacketHandler.GetBytesInFlight()) / float32(cwnd[pathID])
			} else {
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

	if maxRTT[firstPath] <= maxRTT[secondPath] {
		maxRTT[firstPath] = maxRTT[secondPath]
	} else {
		maxRTT[secondPath] = maxRTT[firstPath]
	}

	if float32(packetNumber[firstPath]) > 0 {
		reWard[firstPath] = (1-s.scheduler.Alpha-s.scheduler.Beta)*float64(cwndlevel[firstPath]) - s.scheduler.Alpha*float64(NormalizeTimes(sRTT[firstPath])/50) - s.scheduler.Beta*float64(5*float32(retransNumber[firstPath])/float32(packetNumber[firstPath]))
	} else {
		reWard[firstPath] = (1-s.scheduler.Alpha-s.scheduler.Beta)*float64(cwndlevel[firstPath]) - s.scheduler.Alpha*float64(NormalizeTimes(sRTT[firstPath])/50)
	}
	if float32(packetNumber[secondPath]) > 0 {
		reWard[secondPath] = (1-s.scheduler.Alpha-s.scheduler.Beta)*float64(cwndlevel[secondPath]) - s.scheduler.Alpha*float64(NormalizeTimes(sRTT[secondPath])/50) - s.scheduler.Beta*float64(5*float32(retransNumber[secondPath])/float32(packetNumber[secondPath]))
	} else {
		reWard[secondPath] = (1-s.scheduler.Alpha-s.scheduler.Beta)*float64(cwndlevel[secondPath]) - s.scheduler.Alpha*float64(NormalizeTimes(sRTT[secondPath])/50)
	}

	//State
	oldBSend := s.scheduler.QoldState[State{id: pth.pathID, pktnumber: rcvdpacketNumber, prob: 1}]
	delete(s.scheduler.QoldState, State{id: pth.pathID, pktnumber: rcvdpacketNumber, prob: 1})

	var BSend protocol.ByteCount
	var BSend1 float32

	BSend, _ = s.flowControlManager.SendWindowSize(protocol.StreamID(5))
	BSend1 = float32(BSend) / (float32(protocol.DefaultMaxCongestionWindow) * 300)
	var ro, col, ro1 int8

	if pth.pathID == firstPath {
		col = 0
		if reWard[firstPath] == 0 {
			return
		}
	} else {
		if reWard[secondPath] == 0 {
			return
		}
		col = 1
	}

	if float64(BSend1) < sch.Qstate[0] {
		ro = 0
	} else if float64(BSend1) < sch.Qstate[1] {
		ro = 1
	} else if float64(BSend1) < sch.Qstate[2] {
		ro = 2
	} else if float64(BSend1) < sch.Qstate[3] {
		ro = 3
	} else if float64(BSend1) < sch.Qstate[4] {
		ro = 4
	} else if float64(BSend1) < sch.Qstate[5] {
		ro = 5
	} else if float64(BSend1) < sch.Qstate[6] {
		ro = 6
	} else {
		ro = 7
	}

	if float64(oldBSend) < sch.Qstate[0] {
		ro1 = 0
	} else if float64(oldBSend) < sch.Qstate[1] {
		ro1 = 1
	} else if float64(oldBSend) < sch.Qstate[2] {
		ro1 = 2
	} else if float64(oldBSend) < sch.Qstate[3] {
		ro1 = 3
	} else if float64(oldBSend) < sch.Qstate[4] {
		ro1 = 4
	} else if float64(oldBSend) < sch.Qstate[5] {
		ro1 = 5
	} else if float64(oldBSend) < sch.Qstate[6] {
		ro1 = 6
	} else {
		ro1 = 7
	}

	var maxNextState float64
	if s.scheduler.Qqtable[Store{Row: ro, Col: 1}] > s.scheduler.Qqtable[Store{Row: ro, Col: 0}] {
		maxNextState = s.scheduler.Qqtable[Store{Row: ro, Col: 1}]
	} else {
		maxNextState = s.scheduler.Qqtable[Store{Row: ro, Col: 0}]
	}

	newValue := (1-s.scheduler.Delta)*s.scheduler.Qqtable[Store{Row: ro1, Col: col}] + s.scheduler.Delta*(reWard[pth.pathID]+s.scheduler.Gamma*maxNextState)

	s.scheduler.Qqtable[Store{Row: ro1, Col: col}] = newValue
}
