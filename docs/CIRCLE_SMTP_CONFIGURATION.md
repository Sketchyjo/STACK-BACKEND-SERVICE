# Circle User-Controlled Wallet SMTP Configuration Guide

This guide helps you configure SMTP settings for Circle's user-controlled wallet integration.

## Required SMTP Settings for Circle Console

### Option 1: Using Resend (Recommended for your setup)

Based on your current configuration, you can use Resend's SMTP service:

```
From: no-reply@stackservice.com (or your verified domain)
Host: smtp.resend.com
Port: 587
Username: resend
Password: re_fdRkJf4D_5MmAsWiQkXa34HwKqvoMLSYT (your Resend API key)
```

### Option 2: Using SendGrid

If you prefer SendGrid:

```
From: no-reply@yourdomain.com
Host: smtp.sendgrid.net
Port: 587
Username: apikey
Password: SG.your-sendgrid-api-key
```

### Option 3: Using Gmail SMTP

For development/testing:

```
From: your-gmail@gmail.com
Host: smtp.gmail.com
Port: 587
Username: your-gmail@gmail.com
Password: your-app-specific-password (not regular password)
```

## OTP Email Customization Settings

### Subject Line
```
{{{code}}} is your login code for Stacks
```

### Email Template
Use the default HTML template provided by Circle, but ensure:
- The `{{code}}` variable is included for OTP display
- Replace "YOUR LOGO" with your company logo
- Customize colors and styling to match your brand
- Include your company name in the template

## Environment Variables Setup

Add these to your `.env` file:

```bash
# SMTP Configuration for Circle
SMTP_HOST=smtp.resend.com
SMTP_PORT=587
SMTP_USERNAME=resend
SMTP_PASSWORD=re_your-resend-api-key
SMTP_FROM_EMAIL=no-reply@yourdomain.com

# Email Configuration
EMAIL_PROVIDER=resend
EMAIL_API_KEY=re_your-resend-api-key
EMAIL_FROM_EMAIL=no-reply@yourdomain.com
EMAIL_FROM_NAME=Your Company Name
EMAIL_REPLY_TO=support@yourdomain.com
```

## Domain Verification

Before using email services:

1. **For Resend**: Verify your domain in the Resend dashboard
2. **For SendGrid**: Verify your domain in SendGrid settings
3. **For Gmail**: Use app-specific passwords, not your regular password

## Testing the Configuration

After setting up:

1. Use Circle's "Test Email" feature in the console
2. Verify emails are being sent successfully
3. Check that OTP codes are properly formatted in emails
4. Test the complete user flow from signup to wallet creation

## Security Considerations

- Never commit API keys to version control
- Use environment variables for all sensitive data
- Rotate API keys regularly
- Monitor email sending limits and quotas
- Set up proper SPF, DKIM, and DMARC records for your domain

## Troubleshooting

### Common Issues:

1. **Authentication Failed**: Check username/password combination
2. **Connection Timeout**: Verify SMTP host and port
3. **Emails Not Delivered**: Check spam folders and domain reputation
4. **OTP Not Displaying**: Ensure `{{code}}` variable is in email template

### Debug Steps:

1. Check Circle Console logs for SMTP errors
2. Verify domain verification status
3. Test SMTP connection with tools like `telnet`
4. Check email service provider logs
