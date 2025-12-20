import { defineStore } from 'pinia';

export const useRoomStore = defineStore('room', {
  state: () => ({
    roomId: '',
    token: '',
    username: ''
  }),
  actions: {
    setSession(payload: { roomId: string; token: string; username: string }) {
      this.roomId = payload.roomId;
      this.token = payload.token;
      this.username = payload.username;
    }
  }
});
