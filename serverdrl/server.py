from flask import Flask, request, jsonify
from model.sac.main import create_model, fetch_replay_buffer

app = Flask(__name__)

models = {}

@app.route('/set_model', methods=['POST'])
def set_model():
    data = request.json
    state_dim = data['state_dim']
    action_dim = data['action_dim']
    max_action = data['max_action']
    connection_id = data['connection_id']

    models[connection_id] = create_model(state_dim, action_dim, max_action)
    
    return jsonify({"status": "Model created", "connection_id": connection_id})

@app.route('/select_action', methods=['POST'])
def select_action():
    data = request.json
    state = data['state']
    connection_id = data['connection_id']

    model = models.get(connection_id)
    if model is None:
        return jsonify({"error": "Model not found"}), 404

    action_probs = model.select_action_prob(state)
    return jsonify({"action_probs": action_probs.tolist()})

@app.route('/train_model', methods=['POST'])
def train_model():
    data = request.json
    replay_buffer = data['replay_buffer']
    iterations = data['iterations']
    connection_id = data['connection_id']

    model = models.get(connection_id)
    if model is None:
        return jsonify({"error": "Model not found"}), 404

    model.train(replay_buffer, iterations)
    model.plot_training_history(connection_id)

    return jsonify({"status": "Model trained"})

@app.route('/plot_training_history', methods=['POST'])
def plot_training_history():
    data = request.json
    connection_id = data['connection_id']

    model = models.get(connection_id)
    if model is None:
        return jsonify({"error": "Model not found"}), 404

    model.plot_training_history()
    return jsonify({"status": "Plot generated"})

if __name__ == '__main__':
    app.run(debug=True)
