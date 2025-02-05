#!/bin/bash
#!/bin/sleep
#!/bin/sh

export PATH=$PATH:/usr/local/go/bin
sudo mn -c

cd /home/"$(whoami)"/go/src/github.com/lucas-clemente/quic-go ; go build ; go install ./...
cd /home/"$(whoami)"/go/src/github.com/lucas-clemente/simulationFileServer
pwd
cp /home/"$(whoami)"/go/bin/example ./serverMPQUIC
cp /home/"$(whoami)"/go/bin/client_benchmarker ./clientMPQUIC

sudo rm ./logs/*
# sudo rm ./output/result-wireless/*
for i in {3..5}
do
    cp "clientMPQUIC" "clientMPQUIC$i"
done

echo -e "0.3\n0.5\n0.5\n0.4\n0.4\n0.0" > ./config/qsat 
echo -e "12\n0.3" > ./config/sac 

declare -a mdlArr=("none") #("none" "dif1" "dif2" "mob1" "gm2d" "rwp2d" "rpgm2d" "gm3d" "rwp3d" "rpgm3d")
declare -a numArr=("1") #("1" "3" "5")
declare -a filArr=("1MB") 
declare -a bgrArr=("0")
declare -a frqArr=("0")
declare -a bwdArr=("25") #bandwidth
declare -a owdArr=("10")  #one-way delay
declare -a varArr=("8")
declare -a losArr=("1.5")

# declare -a schArr=("blest" "ecf" "peek" "rtt" "fuzzyqsat" "sac")
declare -a schArr=("rtt")
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
                                        sudo rm ./logs/*
                                        sudo -E env "PATH=$PATH" python wifi_scenario3.py --model ${mdl} --client ${num} --file ${fil} --background ${bgr} --frequency ${frq} --bandwidth ${bwd} --delay ${owd} --variation ${var} --loss ${los} --scheduler ${sch} 
                                        # sudo mv ./logs/server.logs ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-server.logs
                                        # sudo mv ./logs/client.logs ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-client.logs
                                        # sudo mv ./logs/result3.csv ./output/${mdl}-${num}-${fil}-${bgr}-${frq}-${bwd}-${owd}-${var}-${los}-${sch}-result3.csv   

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
