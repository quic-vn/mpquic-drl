#!/bin/bash
#!/bin/sleep
#!/bin/sh

sudo mn -c

cd /home/"$(whoami)"/go/src/github.com/lucas-clemente/quic-go ; go build ; go install ./...
cd /home/"$(whoami)"/go/src/github.com/lucas-clemente/simulationFileServer
pwd
cp /home/"$(whoami)"/go/bin/example ./serverMPQUIC
cp /home/"$(whoami)"/go/bin/client_benchmarker ./clientMPQUIC

sudo rm ./logs/*
# sudo rm ./output/result-wireless/*
for i in {3..10}
do
    cp "clientMPQUIC" "clientMPQUIC$i"
done

declare -a mdlArr=("none") #("none" "dif1" "dif2" "mobi")
declare -a numArr=("1") #("1" "4" "8")
declare -a filArr=("1MB") 
declare -a bgrArr=("0")
declare -a frqArr=("0")
declare -a bwdArr=("30") #bandwidth
declare -a owdArr=("10") #one-way delay
declare -a varArr=("10") #variation delay
declare -a losArr=("0") #pkt loss 
# declare -a schArr=("sacrx" "sacmulti" "random" "rtt" "peek" "multiclients"  "sacmultiJoinCC")
declare -a schArr=("sac")

for mdl in "${mdlArr[@]}"
do 
    for num in "${numArr[@]}"
    do 
        for fil in "${filArr[@]}"
        do
            for bgr in "${bgrArr[@]}"
            do 
                for frq in "${frqArr[@]}"
                do 
                    for bwd in "${bwdArr[@]}"
                    do 
                        for owd in "${owdArr[@]}"
                        do
                            for var in "${varArr[@]}"
                            do
                                for los in "${losArr[@]}"
                                do 
                                    for sch in "${schArr[@]}"
                                    do
                                        echo "$mdl-$num-$fil-$bgr-$frq-$bwd-$owd-$var-$los-$sch"
                                        sudo -E env "PATH=$PATH" python wifi_scenario.py --model ${mdl} --client ${num} --file ${fil} --background ${bgr} --frequency ${frq} --bandwidth ${bwd} --delay ${owd} --variation ${var} --loss ${los} --scheduler ${sch} 
                                        sudo mv ./logs/server.logs ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-server.logs
                                        sudo mv ./logs/client.logs ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-client.logs
                                        sudo mv ./logs/result3.csv ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-result3.csv   
                                        sudo mv ./logs/result4.csv ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-result4.csv   
                                        sudo mv ./logs/result5.csv ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-result5.csv   
                                        sudo mv ./logs/result6.csv ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-result6.csv   
                                        sudo mv ./logs/result7.csv ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-result7.csv   
                                        sudo mv ./logs/result8.csv ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-result8.csv   
                                        sudo mv ./logs/result9.csv ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-result9.csv   
                                        sudo mv ./logs/result10.csv ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-result10.csv   

                                        sudo mv ./logs/server-flask.logs ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-flask.logs  
                                        sudo mv ./logs/training_history.png ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-training.png 
                                        sudo mv ./logs/training_history_3.png ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-training_3.png 
                                        sudo mv ./logs/training_history_4.png ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-training_4.png 
                                        sudo mv ./logs/training_history_5.png ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-training_5.png 
                                        sudo mv ./logs/training_history_6.png ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-training_6.png 
                                        sudo mv ./logs/training_history_7.png ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-training_7.png 
                                        sudo mv ./logs/training_history_8.png ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-training_8.png 
                                        sudo mv ./logs/training_history_9.png ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-training_9.png 
                                        sudo mv ./logs/training_history_10.png ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-training_10.png 

                                        sudo mn -c
                                        sleep 10
                                    done
                                done
                            done
                        done
                    done
                done
            done
        done
    done
done
