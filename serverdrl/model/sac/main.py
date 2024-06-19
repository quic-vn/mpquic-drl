import torch
import torch.nn as nn
import torch.optim as optim
import numpy as np
import matplotlib.pyplot as plt
import random

class Actor(nn.Module):
    def __init__(self, state_dim, action_dim, max_action):
        super(Actor, self).__init__()
        self.l1 = nn.Linear(state_dim, 400)
        self.l2 = nn.Linear(400, 300)
        self.l3 = nn.Linear(300, action_dim)

    def forward(self, state):
        a = torch.relu(self.l1(state))
        a = torch.relu(self.l2(a))
        return torch.softmax(self.l3(a), dim=-1)

class Critic(nn.Module):
    def __init__(self, state_dim, action_dim):
        super(Critic, self).__init__()
        self.l1 = nn.Linear(state_dim + action_dim, 400)
        self.l2 = nn.Linear(400, 300)
        self.l3 = nn.Linear(300, 1)

    def forward(self, state, action):
        q = torch.relu(self.l1(torch.cat([state, action], 1)))
        q = torch.relu(self.l2(q))
        return self.l3(q)

class SACModel:
    def __init__(self, state_dim, action_dim, max_action):
        self.actor = Actor(state_dim, action_dim, max_action)
        self.critic = Critic(state_dim, action_dim)
        self.critic_target = Critic(state_dim, action_dim)
        self.critic_target.load_state_dict(self.critic.state_dict())
        self.actor_optimizer = optim.Adam(self.actor.parameters())
        self.critic_optimizer = optim.Adam(self.critic.parameters())
        self.max_action = max_action
        self.training_history = {"critic_loss": [], "actor_loss": [], "rewards": []}

    def train(self, replay_buffer, iterations):
        state, action, reward, next_state, done = zip(*[(torch.FloatTensor(exp['state']),
                                                            torch.FloatTensor([exp['action']]),
                                                            torch.FloatTensor([exp['reward']]),
                                                            torch.FloatTensor(exp['next_state']),
                                                            torch.FloatTensor([exp['done']])) for exp in replay_buffer])
        state = torch.stack(state)
        action = torch.stack(action)
        reward = torch.stack(reward)
        next_state = torch.stack(next_state)
        done = torch.stack(done)

        with torch.no_grad():
            next_action = self.actor(next_state)
            target_q = reward + (1 - done) * self.critic_target(next_state, next_action)

        current_q = self.critic(state, action)
        critic_loss = nn.MSELoss()(current_q, target_q)

        self.critic_optimizer.zero_grad()
        critic_loss.backward()
        self.critic_optimizer.step()

        actor_loss = -self.critic(state, self.actor(state)).mean()

        self.actor_optimizer.zero_grad()
        actor_loss.backward()
        self.actor_optimizer.step()

        self.training_history["critic_loss"].append(critic_loss.item())
        self.training_history["actor_loss"].append(actor_loss.item())
        self.training_history["rewards"].append(reward.mean().item())

    def select_action_prob(self, state):
        state = torch.FloatTensor(state).unsqueeze(0)
        action_probs = self.actor(state).cpu().data.numpy().flatten()
        return action_probs

    def plot_training_history(self, modelid):
        plt.figure(figsize=(12, 6))

        plt.subplot(1, 3, 1)
        plt.plot(self.training_history["critic_loss"], label="Critic Loss")
        plt.xlabel("Iterations")
        plt.ylabel("Loss")
        plt.title("Critic Loss")
        plt.legend()

        plt.subplot(1, 3, 2)
        plt.plot(self.training_history["actor_loss"], label="Actor Loss")
        plt.xlabel("Iterations")
        plt.ylabel("Loss")
        plt.title("Actor Loss")
        plt.legend()

        plt.subplot(1, 3, 3)
        plt.plot(self.training_history["rewards"], label="Rewards")
        plt.xlabel("Iterations")
        plt.ylabel("Reward")
        plt.title("Rewards")
        plt.legend()

        # plt.tight_layout()
        # plt.show()
        plt.tight_layout()
        plt.savefig(f'logs/training_history_{modelid}.png')
        plt.close()

def create_model(state_dim, action_dim, max_action):
    return SACModel(state_dim, action_dim, max_action)

class ReplayBuffer:
    def __init__(self):
        self.buffer = []

    def sample(self):
        state = np.random.rand(1, 6)
        action = np.random.rand(1, 3)
        reward = np.random.rand(1, 1)
        next_state = np.random.rand(1, 6)
        done = np.random.randint(0, 2, size=(1, 1))
        return state, action, reward, next_state, done

def fetch_replay_buffer(replay_buffer_id):
    return ReplayBuffer()
