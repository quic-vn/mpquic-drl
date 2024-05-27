from flask import Flask, request, jsonify
from model.sac.main import Environment as SACEnv
import matplotlib.pyplot as plt

app = Flask(__name__)

# Initialize the SAC environment
sac_env = SACEnv()

# Set the default environment
current_env = sac_env

@app.route('/set_model', methods=['POST'])
def set_model():
    global current_env
    model_type = request.json.get('model_type', 'sac')
    
    if model_type == 'sac':
        current_env = sac_env
    else:
        return jsonify({'error': 'Invalid model type'}), 400
    
    return jsonify({'status': f'Model set to {model_type}'}), 200

@app.route('/flag_training', methods=['POST'])
def flag_training():
    print("TrANDA")
    try:
        current_env.agent.train(batch_size=64)
        return jsonify({'status': 'Training'}), 200
    except Exception as e:
        print(f"Error in training: {e}")
        return jsonify({'error': str(e)}), 500

@app.route('/get_action', methods=['POST'])
def get_action():
    try:
        # Retrieve and validate the state from the request
        state_json = request.json.get('state')
        if not state_json:
            return jsonify({'error': 'State data is missing'}), 400

        required_fields = ['CWNDf', 'INPf', 'SRTTf', 'CWNDs', 'INPs', 'SRTTs']
        for field in required_fields:
            if field not in state_json:
                return jsonify({'error': f'Missing field in state data: {field}'}), 400

        # Convert state data to list
        state = [state_json['CWNDf'], state_json['INPf'], state_json['SRTTf'], 
                 state_json['CWNDs'], state_json['INPs'], state_json['SRTTs']]
        # print(f"Received state: {state}")

        # Get action probability from the SAC agent
        prob = current_env.agent.get_action_probability(state)
        # print(f"Action probability: {prob}")

        # Return the action probability as a response
        return jsonify({'probability': prob.tolist()}), 200

    except Exception as e:
        # Log and return any errors that occur
        print(f"Error in get_action: {e}")
        return jsonify({'error': str(e)}), 500

@app.route('/update_reward', methods=['POST'])
def update_reward():
    try:
        data = request.json
        state_json = data['state']
        next_state_json = data['next_state']
        
        state = [state_json['CWNDf'], state_json['INPf'], state_json['SRTTf'], 
                 state_json['CWNDs'], state_json['INPs'], state_json['SRTTs']]
        next_state = [next_state_json['CWNDf'], next_state_json['INPf'], next_state_json['SRTTf'], 
                      next_state_json['CWNDs'], next_state_json['INPs'], next_state_json['SRTTs']]
        
        action = data['action']
        reward = data['reward']
        done = data['done']

        current_env.agent.replay_buffer.add(state, action, reward, next_state, done)
        #current_env.agent.add_reward(reward)
        #current_env.agent.train(batch_size=64)

        return jsonify({'status': 'Reward updated'}), 200
    except Exception as e:
        print(f"Error in update_reward: {e}")
        return jsonify({'error': str(e)}), 500

@app.route('/plot_training_history', methods=['GET'])
def plot_training_history():
    try:
        current_env.agent.plot_training_history()
        return jsonify({'status': 'Training history plotted'}), 200
    except Exception as e:
        print(f"Error in plot_training_history: {e}")
        return jsonify({'error': str(e)}), 500

@app.route('/')
def index():
    return "Welcome to the Path Scheduler Training Server!"

@app.route('/status', methods=['GET'])
def status():
    return jsonify({'status': 'Server is running'})

if __name__ == '__main__':
    #app.run(debug=True, host='127.0.0.1', port=8080)
    app.run(debug=True, host='0.0.0.0', port=8080)
