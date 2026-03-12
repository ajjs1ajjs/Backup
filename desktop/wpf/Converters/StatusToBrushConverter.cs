using System;
using System.Globalization;
using System.Windows.Data;
using System.Windows.Media;

namespace NovaBackup.GUI.Converters
{
    public class StatusToBrushConverter : IValueConverter
    {
        public object Convert(object value, Type targetType, object parameter, CultureInfo culture)
        {
            var status = (value as string) ?? string.Empty;
            var key = status.ToLowerInvariant();
            return key switch
            {
                "running" or "in progress" => Brushes.LightGreen,
                "completed" or "success" => Brushes.Green,
                "failed" or "error" => Brushes.Red,
                _ => Brushes.Gray,
            };
        }

        public object ConvertBack(object value, Type targetType, object parameter, CultureInfo culture)
        {
            throw new NotImplementedException();
        }
    }
}
