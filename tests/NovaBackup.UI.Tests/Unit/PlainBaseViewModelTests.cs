using System.ComponentModel;
using System.Runtime.CompilerServices;
using Xunit;
using NovaBackup.GUI.ViewModels;

namespace NovaBackup.UI.Tests.Unit
{
    public class PlainBaseViewModelTests
    {
        [Fact]
        public void BaseViewModelCompat_SetProperty_UpdatesValue()
        {
            var vm = new TestVm();
            vm.Value = 5;
            Assert.Equal(5, vm.Value);
        }
        private class TestVm : BaseViewModelCompat
        {
            public int Value { get => _value; set => SetProperty(ref _value, value); }
            private int _value;
        }
    }
}
