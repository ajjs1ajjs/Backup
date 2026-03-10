import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import apiClient, { CreateJobRequest } from '../api/client';
import './JobForm.css';

interface FormData {
  name: string;
  source: string;
  destination: string;
  provider: string;
  schedule: string;
  enabled: boolean;
}

const initialData: FormData = {
  name: '',
  source: '',
  destination: '',
  provider: 'local',
  schedule: '0 2 * * *',
  enabled: true
};

const providerOptions = [
  { value: 'local', label: 'Local', icon: '&#128190;' },
  { value: 's3', label: 'S3', icon: '&#9729;&#65039;' },
  { value: 'azure', label: 'Azure', icon: '&#128311;' },
  { value: 'gcs', label: 'Google Cloud', icon: '&#128310;' }
];

const schedulePresets = [
  { label: 'Every Hour', value: '0 * * * *' },
  { label: 'Daily at 2 AM', value: '0 2 * * *' },
  { label: 'Daily at 3 AM', value: '0 3 * * *' },
  { label: 'Weekly (Sunday)', value: '0 3 * * 0' },
  { label: 'Monthly (1st)', value: '0 3 1 * *' }
];

const JobForm: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEditMode = !!id;

  const [formData, setFormData] = useState<FormData>(initialData);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (isEditMode && id) {
      setLoading(true);
      apiClient.getJob(parseInt(id, 10))
        .then(job => {
          setFormData({
            name: job.name,
            source: job.source,
            destination: job.destination,
            provider: job.provider,
            schedule: job.schedule,
            enabled: job.enabled
          });
        })
        .catch(err => {
          setError('Failed to load job');
          console.error('Load job error:', err);
        })
        .finally(() => setLoading(false));
    }
  }, [id, isEditMode]);

  const validate = (): boolean => {
    const errors: Record<string, string> = {};
    if (!formData.name.trim()) errors.name = 'Job name is required';
    if (!formData.source.trim()) errors.source = 'Source path is required';
    if (!formData.destination.trim()) errors.destination = 'Destination path is required';
    if (!formData.schedule.trim()) errors.schedule = 'Schedule is required';

    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    try {
      setLoading(true);
      setError(null);

      if (isEditMode && id) {
        await apiClient.updateJob(parseInt(id, 10), formData);
      } else {
        await apiClient.createJob(formData as CreateJobRequest);
      }

      navigate('/jobs');
    } catch (err) {
      setError(isEditMode ? 'Failed to update job' : 'Failed to create job');
      console.error('Save job error:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target;
    const checked = (e.target as HTMLInputElement).checked;

    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }));

    // Clear validation error for this field
    if (validationErrors[name]) {
      setValidationErrors(prev => {
        const newErrors = { ...prev };
        delete newErrors[name];
        return newErrors;
      });
    }
  };

  const handlePresetClick = (value: string) => {
    setFormData(prev => ({ ...prev, schedule: value }));
  };

  if (loading && isEditMode) {
    return (
      <div className="page">
        <div className="loading">Loading job...</div>
      </div>
    );
  }

  return (
    <div className="page">
      <div className="job-form">
        <div className="job-form-header">
          <h2>{isEditMode ? '&#9999;&#65039; Edit Job' : '&#10133; Create Job'}</h2>
          <button className="btn btn-secondary" onClick={() => navigate('/jobs')}>
            &#8592; Back to Jobs
          </button>
        </div>

        {error && <div className="error-banner">{error}</div>}

        <form onSubmit={handleSubmit} className="job-form-content">
          <div className="form-section">
            <h3>&#128203; Basic Information</h3>

            <div className="form-group">
              <label htmlFor="name">Job Name *</label>
              <input
                type="text"
                id="name"
                name="name"
                value={formData.name}
                onChange={handleChange}
                placeholder="e.g., Daily Database Backup"
                className={validationErrors.name ? 'error' : ''}
              />
              {validationErrors.name && (
                <span className="field-error">{validationErrors.name}</span>
              )}
            </div>

            <div className="form-row">
              <div className="form-group">
                <label htmlFor="provider">Storage Provider</label>
                <select
                  id="provider"
                  name="provider"
                  value={formData.provider}
                  onChange={handleChange}
                >
                  {providerOptions.map(opt => (
                    <option key={opt.value} value={opt.value}>
                      {opt.icon} {opt.label}
                    </option>
                  ))}
                </select>
              </div>

              <div className="form-group checkbox-group">
                <label className="checkbox-label">
                  <input
                    type="checkbox"
                    name="enabled"
                    checked={formData.enabled}
                    onChange={handleChange}
                  />
                  <span>Enable Job</span>
                </label>
              </div>
            </div>
          </div>

          <div className="form-section">
            <h3>&#128193; Locations</h3>

            <div className="form-group">
              <label htmlFor="source">Source Path *</label>
              <input
                type="text"
                id="source"
                name="source"
                value={formData.source}
                onChange={handleChange}
                placeholder="e.g., C:\Data or /home/user/documents"
                className={validationErrors.source ? 'error' : ''}
              />
              {validationErrors.source && (
                <span className="field-error">{validationErrors.source}</span>
              )}
            </div>

            <div className="form-group">
              <label htmlFor="destination">Destination Path *</label>
              <input
                type="text"
                id="destination"
                name="destination"
                value={formData.destination}
                onChange={handleChange}
                placeholder="e.g., D:\Backups or /backup/location"
                className={validationErrors.destination ? 'error' : ''}
              />
              {validationErrors.destination && (
                <span className="field-error">{validationErrors.destination}</span>
              )}
            </div>
          </div>

          <div className="form-section">
            <h3>&#9200; Schedule</h3>

            <div className="preset-buttons">
              {schedulePresets.map(preset => (
                <button
                  type="button"
                  key={preset.value}
                  className={`preset-btn ${formData.schedule === preset.value ? 'active' : ''}`}
                  onClick={() => handlePresetClick(preset.value)}
                >
                  {preset.label}
                </button>
              ))}
            </div>

            <div className="form-group">
              <label htmlFor="schedule">Cron Expression *</label>
              <input
                type="text"
                id="schedule"
                name="schedule"
                value={formData.schedule}
                onChange={handleChange}
                placeholder="e.g., 0 2 * * *"
                className={validationErrors.schedule ? 'error' : ''}
              />
              {validationErrors.schedule && (
                <span className="field-error">{validationErrors.schedule}</span>
              )}
              <span className="field-help">
                Format: minute hour day month weekday (e.g., "0 2 * * *" = daily at 2:00 AM)
              </span>
            </div>
          </div>

          <div className="form-actions">
            <button
              type="button"
              className="btn btn-secondary"
              onClick={() => navigate('/jobs')}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={loading}
            >
              {loading ? 'Saving...' : (isEditMode ? 'Update Job' : 'Create Job')}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default JobForm;
