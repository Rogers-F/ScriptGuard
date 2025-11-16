import datetime
import time

print("=" * 50)
print("ScriptGuard 测试脚本")
print("=" * 50)

# 显示当前时间
beijing_tz = datetime.timezone(datetime.timedelta(hours=8))
beijing_time = datetime.datetime.now(beijing_tz)
print(f"\n执行时间: {beijing_time.strftime('%Y-%m-%d %H:%M:%S')} (北京时间)")

# 显示环境信息
import sys
print(f"Python版本: {sys.version}")
print(f"Python路径: {sys.executable}")

# 模拟处理
print("\n开始处理...")
for i in range(5):
    time.sleep(1)
    print(f"进度: {(i+1)*20}%")

print("\n✅ 测试脚本执行成功！")
