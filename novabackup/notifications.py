"""
NovaBackup Notification System

Supports:
- Email notifications (SMTP)
- Telegram bot notifications
- Webhook notifications
- Console logging
"""

import smtplib
import logging
from typing import Dict, Any, Optional, List
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from datetime import datetime
import asyncio
import aiohttp

logger = logging.getLogger("novabackup.notifications")


class NotificationManager:
    """Manages all notification channels."""
    
    def __init__(self):
        self.channels: Dict[str, NotificationChannel] = {}
        self.notification_history: List[Dict[str, Any]] = []
    
    def register_channel(self, name: str, channel: 'NotificationChannel'):
        """Register a notification channel."""
        self.channels[name] = channel
        logger.info(f"Registered notification channel: {name}")
    
    def unregister_channel(self, name: str):
        """Unregister a notification channel."""
        if name in self.channels:
            del self.channels[name]
            logger.info(f"Unregistered notification channel: {name}")
    
    async def send(
        self,
        message: str,
        level: str = "info",
        channels: Optional[List[str]] = None,
        **kwargs
    ):
        """
        Send notification to specified channels.
        
        Args:
            message: Notification message
            level: Notification level (info, warning, error, success)
            channels: List of channel names (if None, send to all)
            **kwargs: Additional arguments for channels
        """
        timestamp = datetime.utcnow().isoformat()
        
        notification = {
            "timestamp": timestamp,
            "message": message,
            "level": level,
            "channels": channels or list(self.channels.keys()),
            "status": "sent",
        }
        
        tasks = []
        target_channels = channels or list(self.channels.keys())
        
        for channel_name in target_channels:
            if channel_name in self.channels:
                channel = self.channels[channel_name]
                tasks.append(self._send_to_channel(channel, message, level, **kwargs))
        
        if tasks:
            results = await asyncio.gather(*tasks, return_exceptions=True)
            
            # Check if any failed
            if any(isinstance(r, Exception) for r in results):
                notification["status"] = "partial_failure"
            elif all(isinstance(r, Exception) for r in results):
                notification["status"] = "failed"
        
        self.notification_history.append(notification)
        
        # Keep only last 1000 notifications
        if len(self.notification_history) > 1000:
            self.notification_history = self.notification_history[-1000:]
        
        logger.info(f"Notification sent: {message[:100]}...")
        return notification
    
    async def _send_to_channel(
        self,
        channel: 'NotificationChannel',
        message: str,
        level: str,
        **kwargs
    ):
        """Send notification to a specific channel."""
        try:
            await channel.send(message, level, **kwargs)
        except Exception as e:
            logger.error(f"Failed to send to {channel.name}: {e}")
            raise
    
    def get_history(self, limit: int = 100) -> List[Dict[str, Any]]:
        """Get recent notification history."""
        return self.notification_history[-limit:]
    
    def clear_history(self):
        """Clear notification history."""
        self.notification_history = []


class NotificationChannel:
    """Base class for notification channels."""
    
    def __init__(self, name: str):
        self.name = name
    
    async def send(self, message: str, level: str, **kwargs):
        """Send notification."""
        raise NotImplementedError


class EmailNotificationChannel(NotificationChannel):
    """Email notification channel via SMTP."""
    
    def __init__(
        self,
        smtp_host: str,
        smtp_port: int,
        smtp_user: str,
        smtp_password: str,
        from_email: str,
        to_emails: List[str],
        use_tls: bool = True,
    ):
        super().__init__("email")
        self.smtp_host = smtp_host
        self.smtp_port = smtp_port
        self.smtp_user = smtp_user
        self.smtp_password = smtp_password
        self.from_email = from_email
        self.to_emails = to_emails
        self.use_tls = use_tls
    
    async def send(self, message: str, level: str, subject: Optional[str] = None, **kwargs):
        """Send email notification."""
        if not subject:
            subject = f"[NovaBackup] {level.upper()}: Backup Notification"
        
        # Create message
        msg = MIMEMultipart()
        msg['From'] = self.from_email
        msg['To'] = ', '.join(self.to_emails)
        msg['Subject'] = subject
        
        # Add timestamp and level to body
        timestamp = datetime.utcnow().strftime("%Y-%m-%d %H:%M:%S UTC")
        body = f"""
        <html>
        <body>
            <h2>NovaBackup Notification</h2>
            <p><strong>Time:</strong> {timestamp}</p>
            <p><strong>Level:</strong> {level.upper()}</p>
            <hr>
            <p>{message}</p>
            <hr>
            <p style="color: #666; font-size: 12px;">
                This is an automated message from NovaBackup.
            </p>
        </body>
        </html>
        """
        
        msg.attach(MIMEText(body, 'html', 'utf-8'))
        
        # Send email
        loop = asyncio.get_event_loop()
        await loop.run_in_executor(
            None,
            self._send_email_sync,
            msg
        )
        
        logger.info(f"Email sent to {self.to_emails}")
    
    def _send_email_sync(self, msg: MIMEMultipart):
        """Synchronous email sending."""
        if self.use_tls:
            server = smtplib.SMTP(self.smtp_host, self.smtp_port)
            server.starttls()
        else:
            server = smtplib.SMTP_SSL(self.smtp_host, self.smtp_port)
        
        server.login(self.smtp_user, self.smtp_password)
        server.send_message(msg)
        server.quit()


class TelegramNotificationChannel(NotificationChannel):
    """Telegram bot notification channel."""
    
    def __init__(
        self,
        bot_token: str,
        chat_ids: List[str],
    ):
        super().__init__("telegram")
        self.bot_token = bot_token
        self.chat_ids = chat_ids
        self.api_url = f"https://api.telegram.org/bot{bot_token}/sendMessage"
    
    async def send(self, message: str, level: str, **kwargs):
        """Send Telegram notification."""
        # Format message with emoji based on level
        emojis = {
            "info": "ℹ️",
            "success": "✅",
            "warning": "⚠️",
            "error": "❌",
        }
        emoji = emojis.get(level, "📢")
        
        formatted_message = f"{emoji} *NovaBackup Notification*\n\n{message}"
        
        async with aiohttp.ClientSession() as session:
            tasks = []
            for chat_id in self.chat_ids:
                task = self._send_to_chat(session, chat_id, formatted_message)
                tasks.append(task)
            
            await asyncio.gather(*tasks, return_exceptions=True)
    
    async def _send_to_chat(self, session: aiohttp.ClientSession, chat_id: str, message: str):
        """Send message to a specific chat."""
        payload = {
            "chat_id": chat_id,
            "text": message,
            "parse_mode": "Markdown",
        }
        
        async with session.post(self.api_url, json=payload) as response:
            if response.status != 200:
                logger.error(f"Telegram API error: {response.status}")
            else:
                logger.info(f"Telegram sent to {chat_id}")


class WebhookNotificationChannel(NotificationChannel):
    """Webhook notification channel (for custom integrations)."""
    
    def __init__(
        self,
        webhook_url: str,
        headers: Optional[Dict[str, str]] = None,
    ):
        super().__init__("webhook")
        self.webhook_url = webhook_url
        self.headers = headers or {}
    
    async def send(self, message: str, level: str, **kwargs):
        """Send webhook notification."""
        payload = {
            "timestamp": datetime.utcnow().isoformat(),
            "level": level,
            "message": message,
            "source": "NovaBackup",
            **kwargs
        }
        
        async with aiohttp.ClientSession() as session:
            async with session.post(
                self.webhook_url,
                json=payload,
                headers=self.headers
            ) as response:
                if response.status not in [200, 201, 202, 204]:
                    logger.error(f"Webhook error: {response.status}")
                else:
                    logger.info(f"Webhook notification sent")


class ConsoleNotificationChannel(NotificationChannel):
    """Console notification channel (for development)."""
    
    def __init__(self):
        super().__init__("console")
    
    async def send(self, message: str, level: str, **kwargs):
        """Print to console."""
        colors = {
            "info": "\033[94m",
            "success": "\033[92m",
            "warning": "\033[93m",
            "error": "\033[91m",
        }
        reset = "\033[0m"
        color = colors.get(level, "")
        
        timestamp = datetime.utcnow().strftime("%Y-%m-%d %H:%M:%S")
        print(f"{color}[{timestamp}] [{level.upper()}] {message}{reset}")


# Global notification manager instance
_notification_manager: Optional[NotificationManager] = None


def get_notification_manager() -> NotificationManager:
    """Get or create global notification manager."""
    global _notification_manager
    if _notification_manager is None:
        _notification_manager = NotificationManager()
    return _notification_manager


def setup_notifications_from_env():
    """Setup notification channels from environment variables."""
    import os
    
    manager = get_notification_manager()
    
    # Email notifications
    smtp_host = os.getenv("NOVABACKUP_SMTP_HOST")
    smtp_port = int(os.getenv("NOVABACKUP_SMTP_PORT", "587"))
    smtp_user = os.getenv("NOVABACKUP_SMTP_USER")
    smtp_password = os.getenv("NOVABACKUP_SMTP_PASSWORD")
    smtp_from = os.getenv("NOVABACKUP_SMTP_FROM")
    smtp_to = os.getenv("NOVABACKUP_SMTP_TO", "")
    
    if smtp_host and smtp_user and smtp_password:
        email_channel = EmailNotificationChannel(
            smtp_host=smtp_host,
            smtp_port=smtp_port,
            smtp_user=smtp_user,
            smtp_password=smtp_password,
            from_email=smtp_from or smtp_user,
            to_emails=[email.strip() for email in smtp_to.split(",") if email.strip()],
        )
        manager.register_channel("email", email_channel)
        logger.info("Email notifications enabled")
    
    # Telegram notifications
    telegram_token = os.getenv("NOVABACKUP_TELEGRAM_BOT_TOKEN")
    telegram_chat_ids = os.getenv("NOVABACKUP_TELEGRAM_CHAT_IDS", "")
    
    if telegram_token and telegram_chat_ids:
        telegram_channel = TelegramNotificationChannel(
            bot_token=telegram_token,
            chat_ids=[id.strip() for id in telegram_chat_ids.split(",") if id.strip()],
        )
        manager.register_channel("telegram", telegram_channel)
        logger.info("Telegram notifications enabled")
    
    # Webhook notifications
    webhook_url = os.getenv("NOVABACKUP_WEBHOOK_URL")
    if webhook_url:
        webhook_headers = {}
        if os.getenv("NOVABACKUP_WEBHOOK_AUTH_TOKEN"):
            webhook_headers["Authorization"] = f"Bearer {os.getenv('NOVABACKUP_WEBHOOK_AUTH_TOKEN')}"
        
        webhook_channel = WebhookNotificationChannel(
            webhook_url=webhook_url,
            headers=webhook_headers,
        )
        manager.register_channel("webhook", webhook_channel)
        logger.info("Webhook notifications enabled")
    
    # Console notifications (development)
    if os.getenv("NOVABACKUP_DEBUG", "").lower() == "true":
        console_channel = ConsoleNotificationChannel()
        manager.register_channel("console", console_channel)
        logger.info("Console notifications enabled (debug mode)")
    
    return manager


async def notify(
    message: str,
    level: str = "info",
    channels: Optional[List[str]] = None,
    **kwargs
):
    """
    Send notification via global manager.
    
    Usage:
        await notify("Backup completed successfully", level="success")
        await notify("Backup failed!", level="error", channels=["telegram", "email"])
    """
    manager = get_notification_manager()
    await manager.send(message, level, channels, **kwargs)
