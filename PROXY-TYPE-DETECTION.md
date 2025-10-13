# Proxy Type Detection Logic

## Detection Strategy

The system uses a **multi-layered approach** combining hostname patterns and pricing information to accurately classify proxy types.

## Detection Rules (Priority Order)

### 1. Hostname Pattern (Highest Priority)

**Residential:**
- Hostname starts with `ipv4-` or `ipv6-`
- Example: `ipv4-vt-01.resvn.net`, `ipv6-sg-02.resvn.net`
- **Result**: `residential`

**ISP:**
- Hostname contains "isp" keyword
- Example: `isp-server-01.example.com`
- **Result**: `isp`

**Datacenter:**
- Hostname contains "datacenter", "dc", or "cloud"
- Example: `datacenter-us.example.com`, `dc-proxy-01.net`
- **Result**: `datacenter`

### 2. Price-Based Detection (For Ambiguous Cases)

When hostname doesn't match above patterns, use price to determine type:

| Price Range | Type | Reasoning |
|-------------|------|-----------|
| **≥ 100,000** | `residential` | Premium rotating residential proxies |
| **10,000 - 99,999** | `isp` | Mid-tier ISP or premium static |
| **1 - 9,999** | `static` | Cheap static IP or datacenter |
| **0** | `unknown` | No pricing info available |

### 3. IP Format Detection (Fallback)

If no pattern or price match:
- Format `xxx.xxx.xxx.xxx` → `static`
- Other formats → `unknown`

## Real Examples from CloudMini

### Example 1: Residential (High Price + Pattern)
```json
{
  "ip": "ipv4-vt-01.resvn.net:127",
  "price": 120000,
  "location": "Việt Nam - Viettel"
}
```
**Detection**: `residential`
- ✅ Hostname pattern: `ipv4-*`
- ✅ High price: 120,000 VND
- **Result**: Residential proxy

### Example 2: ISP (Medium Price + IP)
```json
{
  "ip": "160.30.138.137",
  "price": 50000,
  "location": "Singapore"
}
```
**Detection**: `isp`
- ❌ No hostname pattern
- ✅ Medium price: 50,000 VND
- ✅ Raw IP address
- **Result**: ISP proxy (premium static)

### Example 3: Static (Low Price + IP)
```json
{
  "ip": "103.161.178.193",
  "price": 1,
  "location": "Việt Nam - Viettel"
}
```
**Detection**: `static`
- ❌ No hostname pattern
- ✅ Low price: 1 VND (legacy/cheap)
- ✅ Raw IP address
- **Result**: Static IP

## Detection Flow Diagram

```
Input: host + price
         |
         v
┌─────────────────────┐
│ Check hostname      │
│ pattern first       │
└──────┬──────────────┘
       │
       ├─ ipv4-* / ipv6-* ────────> residential
       ├─ contains "isp" ──────────> isp
       ├─ contains "datacenter" ───> datacenter
       │
       v
┌─────────────────────┐
│ No pattern match    │
│ Check price         │
└──────┬──────────────┘
       │
       ├─ price >= 100000 ─────────> residential
       ├─ price >= 10000 ──────────> isp
       ├─ price > 0 & < 10000 ─────> static
       │
       v
┌─────────────────────┐
│ Check IP format     │
└──────┬──────────────┘
       │
       ├─ xxx.xxx.xxx.xxx ─────────> static
       │
       v
     unknown
```

## Code Implementation

### Function: `detectProxyTypeWithPrice(host string, price int) string`

**Location**: `cmd/proxy-fwd/cloudmini.go`

**Usage**:
```go
proxyType := detectProxyTypeWithPrice("103.161.178.193", 1)
// Returns: "static"

proxyType := detectProxyTypeWithPrice("ipv4-vt-01.resvn.net", 120000)
// Returns: "residential"

proxyType := detectProxyTypeWithPrice("160.30.138.137", 50000)
// Returns: "isp"
```

## UI Display

Each proxy type has a distinct badge:

| Type | Icon | Color | Badge |
|------|------|-------|-------|
| Residential | 🏠 | Green | `bg-green-100 text-green-800` |
| ISP | 📡 | Blue | `bg-blue-100 text-blue-800` |
| Static | 🔒 | Gray | `bg-gray-100 text-gray-800` |
| Datacenter | 🏢 | Purple | `bg-purple-100 text-purple-800` |
| Unknown | ❓ | Yellow | `bg-yellow-100 text-yellow-800` |

## Price Thresholds (CloudMini VND)

Based on observed pricing patterns:

```
Premium Residential: 120,000 VND/month
Standard ISP:         50,000 VND/month
Budget Static:            1 VND (legacy)
```

## Manual Override

If detection is incorrect, types are stored in `proxies.yaml` and can be manually edited:

```yaml
items:
  - id: 103-161-178-193-39998
    host: 103.161.178.193
    port: 39998
    proxy_type: static  # ← Can be manually changed
    status: stopped
```

## Testing Detection

Use the Pool tab to verify:
1. Sync proxies from CloudMini
2. Check Type column shows correct badges
3. Filter by type to verify classification
4. Compare with price/hostname patterns

## Future Improvements

Potential enhancements:
- Use `location` field for geo-based classification
- Track proxy performance by type
- Add custom price thresholds per provider
- Support multiple proxy providers beyond CloudMini
