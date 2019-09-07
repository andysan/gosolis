
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

| ID     | Standard      |
| ------ | ------------- |
| 0x0000 | Generating    |
| 0x0001 | Generating    |
| 0x0002 | Low wind/sun  |
| 0x0003 | Initializing  |
| 0x1010 | OV-G-V        |
| 0x1011 | UN-G-V        |
| 0x1012 | OV-G-F        |
| 0x1013 | UN-G-F        |
| 0x1014 | G-IMP         |
| 0x1015 | NO-G          |
| 0x1016 | G-PHASE       |
| 0x1017 | G-F-FLU       |
| 0x1018 | OV-G-I        |
| 0x1020 | OV-DC         |
| 0x1021 | OV-BUS        |
| 0x1022 | UNB-BUS       |
| 0x1023 | UN_BUS        |
| 0x1024 | UNB2_BUS      |
| 0x1025 | OV-DCA-I      |
| 0x1026 | OV-DCB-I      |
| 0x1030 | GRID-INTF     |
| 0x1031 | INI-FAULT     |
| 0x1032 | OV-TEM        |
| 0x1033 | GROUND-FAULT  |
| 0x1034 | ILeak-FAULT   |
| 0x1035 | Relay-FAULT   |
| 0x1036 | DSP-B-FAULT   |
| 0x1037 | DCInj-FAULT   |
| 0x1038 | 12Power-FAULT |
| 0x1039 | ILeak-Check   |
| 0x1040 | AFCI-Check    |
| 0x1041 | AFCI-FAULT    |


# Common datastructures

## Inverter info

    +------+------+--------+--------+-------+----------+-------+-------+
    |  0-1 | 2-3  |  4-5   |  6-7   | 8-9   |  10-13   | 14-15 | 16-17 |
    +------+------+--------+--------+-------+----------+-------+-------+
    | V in | I in | V grid | I grid | Temp. | Tot. kWh | State | Err.  |
    +------+------+--------+--------+-------+----------+-------+-------+

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

    +------------+---------+-----------+
    |   37-44    |  45-47  |   48-49   |
    +------------+---------+-----------+
    | Serial no. | Unknown | Ext. ver? |
    +------------+---------+-----------+

## Interface Status

    +--------------------+------+-------+------+--------+----------+
    |         0-31       |  32  |  33   |  34  | 35-36  |   37-39  |
    +--------------------+------+-------+------+--------+----------+
    |  Null-term. string | RSSI | Conn. | 0x01 | Status | 0x000000 |
    +--------------------+------+-------+------+--------+----------+

String message: Null terminated(?) string with inverted SN and
interface IP.

Conn.: Some type of connection status flag (seems to be a Boolean).


Status: Bit field (big endian)
0x0008: No IP
0x0010: Inverse of status bool

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
| 0xc1 | Interface status   |

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

## Interface Status (0xc1)

Send log message to inverter. Observed from the official WiFi
interface. Does not seem to have any effect on some inverters.

Parameters: Interface status message padded to 40 bytes.
Returns: Ack.
