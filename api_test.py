import requests
import json
import os

# 默认基础 URL
DEFAULT_BASE_URL = "http://localhost:8080"
# 您可以通过环境变量 API_BASE_URL 来覆盖基础 URL
BASE_URL = os.environ.get("API_BASE_URL", DEFAULT_BASE_URL)

def print_request_info(method, url, payload=None):
    """打印请求信息"""
    print(f"--- Testing {method} {url} ---")
    if payload:
        print("Request Body:")
        try:
            # 使用 ensure_ascii=False 来正确显示中文字符
            print(json.dumps(payload, indent=4, ensure_ascii=False))
        except TypeError:
            print("Payload is not JSON serializable.")
            print(payload)


def print_response_info(response):
    """打印响应信息"""
    print(f"Status Code: {response.status_code}")
    try:
        response_json = response.json()
        print("Response JSON:")
        # 使用 ensure_ascii=False 来正确显示中文字符
        print(json.dumps(response_json, indent=4, ensure_ascii=False))
        return response_json # 返回 JSON 以便后续使用
    except json.JSONDecodeError:
        print("Response Content (Not JSON):")
        print(response.text)
        return None # 非 JSON 响应
    except Exception as e:
        print(f"Error processing response: {e}")
        return None
    finally:
        print("-" * 30) # 分隔符


def test_healthz():
    """测试 GET /healthz 接口"""
    url = f"{BASE_URL}/healthz"
    print_request_info("GET", url)
    try:
        response = requests.get(url, timeout=5)
        print_response_info(response)
    except requests.exceptions.RequestException as e:
        print(f"Error during request: {e}")
        print("-" * 30)


def test_ping():
    """测试 GET /api/v1/ping 接口"""
    url = f"{BASE_URL}/api/v1/ping"
    print_request_info("GET", url)
    try:
        response = requests.get(url, timeout=5)
        print_response_info(response)
    except requests.exceptions.RequestException as e:
        print(f"Error during request: {e}")
        print("-" * 30)


def test_search_tickets():
    """测试 POST /api/v1/flights/tickets/search 接口"""
    url = f"{BASE_URL}/api/v1/tc/tickets/search"
    payload = {
        "from": "SHA",
        "to": "PEK",
        "date": "2025-12-01"
    }
    print_request_info("POST", url, payload)
    proxy = {
        "http": "http://127.0.0.1:9000",
        "https": "http://127.0.0.1:9000"
    } # 代理设置
    try:

        response = requests.post(url, json=payload, timeout=10,proxies=proxy) # 增加超时时间
        response_data = print_response_info(response)
        # 尝试从响应中提取 flightId (这依赖于实际的 API 响应结构)
        # 假设响应结构是 {"code": 0, "data": [{"flightId": "some-id", ...}]}
        if response_data and response_data.get("code") == 0:
            flights = response_data.get("data", [])
            if flights and isinstance(flights, list) and len(flights) > 0:
                # 返回第一个航班的 ID，用于后续订单测试
                return flights[0].get("flightId")
    except requests.exceptions.RequestException as e:
        print(f"Error during request: {e}")
        print("-" * 30)
    return None # 如果没有找到 flightId 或请求失败，返回 None


def test_order_ticket(flight_id):
    """
    测试 POST /api/v1/flights/tickets/order 接口。
    注意：需要一个有效的 'flightId'。如果 flight_id 为 None 或无效，则跳过。
    """
    url = f"{BASE_URL}/api/v1/tc/tickets/order"

    if not flight_id:
        print(f"--- Skipping POST {url} ---")
        print("Reason: No valid flightId provided from search results.")
        print("-" * 30)
        return

    payload = {
        "flightId": flight_id,
        "passengers": [
            {
                "name": "测试乘客",
                "idType": "IDCard",
                "idNumber": "110101199003070011" # 示例身份证号
            }
        ],
        "contactName": "测试联系人",
        "contactPhone": "13800138000" # 示例手机号
    }
    print_request_info("POST", url, payload)
    try:
        response = requests.post(url, json=payload, timeout=15) # 增加超时时间
        print_response_info(response)
    except requests.exceptions.RequestException as e:
        print(f"Error during request: {e}")
        print("-" * 30)


if __name__ == "__main__":
    print(f"Starting API tests against base URL: {BASE_URL}")
    print("=" * 30)

    test_healthz()
    test_ping()

    # 执行搜索并尝试获取 flightId
    print("Attempting to search for flights to get a flightId for ordering...")
    found_flight_id = test_search_tickets()

    if found_flight_id:
        print(f"Found flightId: {found_flight_id}. Proceeding to order test.")
    else:
        print("Could not retrieve a flightId from search results. Order test requires manual flightId.")
        # 如果需要，可以在这里设置一个已知的测试 flightId
        # found_flight_id = "MANUAL_TEST_FLIGHT_ID"

    # 使用获取到的或手动的 flightId 进行订单测试
    test_order_ticket(found_flight_id)

    print("=" * 30)
    print("API tests finished.")