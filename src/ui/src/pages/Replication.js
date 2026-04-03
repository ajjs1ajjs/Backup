import React from 'react';
import { Box, Card, CardContent, Typography, Button } from '@mui/material';
import { CloudOff as ReplicateIcon, Add as AddIcon } from '@mui/icons-material';

export default function Replication() {
  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h5" sx={{ fontWeight: 600 }}>Replication</Typography>
        <Button variant="contained" startIcon={<AddIcon />} sx={{ bgcolor: '#4fc3f7', '&:hover': { bgcolor: '#29b6f6' } }}>
          New Replication Job
        </Button>
      </Box>

      <Card sx={{ borderRadius: 2 }}>
        <CardContent sx={{ textAlign: 'center', py: 8 }}>
          <ReplicateIcon sx={{ fontSize: 64, color: '#e0e0e0', mb: 2 }} />
          <Typography variant="h6" color="text.secondary" gutterBottom>Replication not configured</Typography>
          <Typography variant="body2" color="text.secondary" mb={3}>Replicate VMs to a secondary site for disaster recovery</Typography>
          <Button variant="contained" startIcon={<AddIcon />} sx={{ bgcolor: '#4fc3f7' }}>Create Replication Job</Button>
        </CardContent>
      </Card>
    </Box>
  );
}
