import axios, { AxiosInstance } from 'axios';

export interface BackupJob {
  id: number;
  name: string;
  source: string;
  destination: string;
  provider: string;
  schedule: string;
  enabled: boolean;
  status: 'pending' | 'running' | 'completed' | 'failed';
  lastRun?: string;
  nextRun?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Backup {
  id: number;
  jobId: number;
  jobName: string;
  size: number;
  chunkCount: number;
  startTime: string;
  endTime?: string;
  status: 'running' | 'completed' | 'failed';
  error?: string;
}

export interface StorageStats {
  totalSize: number;
  usedSize: number;
  backupCount: number;
  jobCount: number;
}

export interface DashboardStats {
  totalJobs: number;
  activeJobs: number;
  totalBackups: number;
  totalSize: number;
  recentFailures: number;
  lastBackup?: string;
}

export interface CreateJobRequest {
  name: string;
  source: string;
  destination: string;
  provider?: string;
  schedule?: string;
  enabled?: boolean;
}

export interface UpdateJobRequest {
  name?: string;
  source?: string;
  destination?: string;
  provider?: string;
  schedule?: string;
  enabled?: boolean;
}

class ApiClient {
  private client: AxiosInstance;
  private baseURL: string;

  constructor(baseURL: string = 'http://localhost:8080') {
    this.baseURL = baseURL;
    this.client = axios.create({
      baseURL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        console.error('API Error:', error);
        return Promise.reject(error);
      }
    );
  }

  // Dashboard
  async getStats(): Promise<DashboardStats> {
    const response = await this.client.get('/api/v1/stats');
    return response.data;
  }

  // Jobs
  async getJobs(): Promise<BackupJob[]> {
    const response = await this.client.get('/api/v1/jobs');
    return response.data;
  }

  async getJob(id: number): Promise<BackupJob> {
    const response = await this.client.get(`/api/v1/jobs/${id}`);
    return response.data;
  }

  async createJob(job: CreateJobRequest): Promise<BackupJob> {
    const response = await this.client.post('/api/v1/jobs', job);
    return response.data;
  }

  async updateJob(id: number, job: UpdateJobRequest): Promise<BackupJob> {
    const response = await this.client.put(`/api/v1/jobs/${id}`, job);
    return response.data;
  }

  async deleteJob(id: number): Promise<void> {
    await this.client.delete(`/api/v1/jobs/${id}`);
  }

  async toggleJob(id: number, enabled: boolean): Promise<BackupJob> {
    const response = await this.client.patch(`/api/v1/jobs/${id}/toggle`, { enabled });
    return response.data;
  }

  async runJob(id: number): Promise<void> {
    await this.client.post(`/api/v1/jobs/${id}/run`);
  }

  // Backups
  async getBackups(jobId?: number): Promise<Backup[]> {
    const params = jobId ? { jobId } : {};
    const response = await this.client.get('/api/v1/backups', { params });
    return response.data;
  }

  async getBackup(id: number): Promise<Backup> {
    const response = await this.client.get(`/api/v1/backups/${id}`);
    return response.data;
  }

  async restore(backupId: number, destination: string, files?: string[]): Promise<void> {
    await this.client.post('/api/v1/restore', { backupId, destination, files });
  }

  // Storage
  async getStorageStats(): Promise<StorageStats> {
    const response = await this.client.get('/api/v1/storage/stats');
    return response.data;
  }

  // VSS Writers
  async getVSSWriters() {
    const response = await this.client.get('/api/v1/vss/writers');
    return response.data;
  }

  // Credentials
  async getCredentials() {
    const response = await this.client.get('/api/v1/credentials');
    return response.data;
  }

  async createCredential(cred: { name: string; username: string; password: string; domain?: string; type?: string }) {
    const response = await this.client.post('/api/v1/credentials', cred);
    return response.data;
  }

  async deleteCredential(id: string) {
    await this.client.delete(`/api/v1/credentials/${id}`);
  }

  // Replication
  async getReplicationJobs() {
    const response = await this.client.get('/api/v1/replication');
    return response.data;
  }

  async createReplicationJob(job: { source_vm: string; destination_host: string; replication_type: string }) {
    const response = await this.client.post('/api/v1/replication', job);
    return response.data;
  }

  async getReplicationStatus(id: string) {
    const response = await this.client.get(`/api/v1/replication/${id}/status`);
    return response.data;
  }

  async stopReplication(id: string) {
    await this.client.post(`/api/v1/replication/${id}/stop`);
  }

  // Tape
  async getTapeLibraries() {
    const response = await this.client.get('/api/v1/tape/libraries');
    return response.data;
  }

  async getTapeCartridges() {
    const response = await this.client.get('/api/v1/tape/cartridges');
    return response.data;
  }

  async getTapeVaults() {
    const response = await this.client.get('/api/v1/tape/vaults');
    return response.data;
  }

  async createTapeVault(vault: { name: string; description?: string; location?: string; contact?: string }) {
    const response = await this.client.post('/api/v1/tape/vaults', vault);
    return response.data;
  }

  async getTapeJobs() {
    const response = await this.client.get('/api/v1/tape/jobs');
    return response.data;
  }

  async createTapeJob(job: { name: string; source: string; target_vault: string; schedule?: string; retention_days?: number }) {
    const response = await this.client.post('/api/v1/tape/jobs', job);
    return response.data;
  }

  // Settings
  async getSettings() {
    const response = await this.client.get('/api/v1/settings');
    return response.data;
  }

  async updateSettings(settings: any) {
    const response = await this.client.put('/api/v1/settings', settings);
    return response.data;
  }

  // RBAC - Users
  async getUsers() {
    const response = await this.client.get('/api/v1/rbac/users');
    return response.data;
  }

  async createUser(user: { username: string; email: string; password: string }) {
    const response = await this.client.post('/api/v1/rbac/users', user);
    return response.data;
  }

  async updateUser(id: string, user: { email?: string; active?: boolean }) {
    const response = await this.client.put(`/api/v1/rbac/users/${id}`, user);
    return response.data;
  }

  async deleteUser(id: string) {
    await this.client.delete(`/api/v1/rbac/users/${id}`);
  }

  async assignRoleToUser(userId: string, roleId: string) {
    await this.client.post(`/api/v1/rbac/users/${userId}/roles`, { role_id: roleId });
  }

  async removeRoleFromUser(userId: string, roleId: string) {
    await this.client.delete(`/api/v1/rbac/users/${userId}/roles/${roleId}`);
  }

  // RBAC - Roles
  async getRoles() {
    const response = await this.client.get('/api/v1/rbac/roles');
    return response.data;
  }

  async createRole(role: { id: string; name: string; description?: string; permissions?: string[] }) {
    const response = await this.client.post('/api/v1/rbac/roles', role);
    return response.data;
  }

  async updateRole(id: string, role: { name?: string; description?: string; permissions?: string[] }) {
    const response = await this.client.put(`/api/v1/rbac/roles/${id}`, role);
    return response.data;
  }

  async deleteRole(id: string) {
    await this.client.delete(`/api/v1/rbac/roles/${id}`);
  }

  // RBAC - Permissions
  async getPermissions() {
    const response = await this.client.get('/api/v1/rbac/permissions');
    return response.data;
  }

  // Health
  async healthCheck(): Promise<{ status: string; version: string }> {
    const response = await this.client.get('/health');
    return response.data;
  }
}

const apiClient = new ApiClient();
export default apiClient;
