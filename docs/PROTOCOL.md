
# Framing

## Command Framing

    +------+---------+---------+--------+----------+----------+
    |  0   |    1    |    2    |   3    |  4 - 53  |   54     |
    +------+---------+---------+--------+----------+----------+
    | 0x7e | Dev. ID | Command | Length | Data[50] | Checksum |
    +------+---------+---------+--------+----------+----------+

## Acknowledgement Framing

    +------+---------+---------+---+
    |  0   |    1    |    2    | 3 |
    +------+---------+---------+---+
    | 0x7e | Dev. ID | Command | 0 |
    +------+---------+---------+---+

## Checksum

Checksum is calculated as the sum of all bytes in the packet ignoring
the start byte.

# Common constants

## Power standard

| ID   | Standard  |
| ---- | --------- |
| 0x00 | Default?  |
| 0x01 | G59/G83   |
| 0x02 | UL-240V   |
| 0x03 | VDE0126   |
| 0x04 | AS4777    |
| 0x05 | AS4777-NQ |
| 0x06 | CQC       |
| 0x07 | ENEL      |
| 0x08 | UL-208V   |
| 0x09 | MEX-CFE   |
| 0x0A | User def. |
| 0x0B | VDE4105   |
| 0x0C | EN50438DK |
| 0x0D | EN50438IE |
| 0x0E | EN50438NL |
| 0x0F | EN50438T  |
| 0x10 | EN50438L  |

## Inverter state


# Common datastructures

## Inverter info

    +------+------+--------+--------+-------+----------+--------+-------+
    |  0-1 | 2-3  |  4-5   |  6-7   | 8-9   |  10-13   | 14-15  | 16-17 |
    +------+------+--------+--------+-------+----------+--------+-------+
    | V in | I in | V grid | I grid | Temp. | Tot. kWh | Status | Err.  |
    +------+------+--------+--------+-------+----------+--------+-------+

    +---------+-------------+------------+------------+-------------+
    |   18    |     19      |   20-21    |     22     |     23      |
    +---------+-------------+------------+------------+-------------+
    | Product | SW. version | Grid freq. | Power std. | Power curve |
    +---------+-------------+------------+------------+-------------+

    +-------+-------+-------------+-----------+----------+-------+-----------+
    | 24-25 | 26-27 |     28      |   29-30   |  31-32   | 33-34 |  35-36    |
    +-------+-------+-------------+-----------+----------+-------+-----------+
    | V2 in | I2 in | Grid status | Month kWh | Last Mth | Today | Yesterday |
    +-------+-------+-------------+-----------+----------+-------+-----------+

# Commands

| ID   | Command            |
| ---- | -----------------  |
| 0x02 | Grid on            |
| 0x03 | Grid off           |
| 0x05 | Set power standard |
| 0x06 | Ping?              |
| 0xa1 | Get information    |
| 0xa3 | Get power curve    |
| 0xa4 | Select power curve |
| 0xaa | Update power curve |
| 0xc1 | Log message        |

## Grid On (0x02)

Enable grid connection.

Parameters: None
Returns: Ack.

## Grid Off (0x03)

Disable grid connection.

Parameters: None
Returns: Ack.

## Set Power Standard (0x05)

Set the current power standard.

Parameters: 1 byte denoting the standard.
Returns: Ack.


## Ping? (0x06)

Ping or keep alive message.

Parameters: None
Returns: Ack.

## Get Information (0xa1)

Get current inverter status.

Parameters: None
Return: See TODO

## Get Power Curve (0xa3)

TODO

## Select Power Curve (0xa4)

TODO

## Update Power Curve (0xaa)

TODO

## Log Message (0xc1)

Send log message to inverter. Observed from the official WiFi
interface.

Parameters: Null terminated string padded to 40 bytes.
Returns: Ack.
