import React from 'react';
import { Alert, Box, Button, Card, CardContent, Typography } from '@mui/material';
import { Add as AddIcon, CloudOff as ReplicateIcon } from '@mui/icons-material';

export default function Replication() {
  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h5" sx={{ fontWeight: 600 }}>Replication</Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          sx={{ bgcolor: '#4fc3f7', '&:hover': { bgcolor: '#29b6f6' } }}
          disabled
        >
          New Replication Job
        </Button>
      </Box>

      <Alert severity="info" sx={{ mb: 3 }}>
        Replication workflows are not implemented yet in this build. Backup and restore flows are available, but
        cross-site replication still needs a backend pipeline and dedicated APIs.
      </Alert>

      <Card sx={{ borderRadius: 2 }}>
        <CardContent sx={{ textAlign: 'center', py: 8 }}>
          <ReplicateIcon sx={{ fontSize: 64, color: '#e0e0e0', mb: 2 }} />
          <Typography variant="h6" color="text.secondary" gutterBottom>
            Replication not configured
          </Typography>
          <Typography variant="body2" color="text.secondary" mb={3}>
            This area is reserved for future disaster recovery replication jobs.
          </Typography>
          <Button variant="contained" startIcon={<AddIcon />} sx={{ bgcolor: '#4fc3f7' }} disabled>
            Create Replication Job
          </Button>
        </CardContent>
      </Card>
    </Box>
  );
}
