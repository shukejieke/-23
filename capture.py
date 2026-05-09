import argparse
import signal
import sys
import threading
import time
import zlib
from collections import defaultdict
from dataclasses import dataclass
from typing import Dict, List, Set, Tuple, Union

from scapy.all import AsyncSniffer, IP, IPv6, Raw, TCP, get_working_ifaces


@dataclass
class LossState:
    cmd_id: int
    needed_buf_size: int
    payload: bytes


class BufferReader:
    def __init__(self, data: bytes):
        self.data = data
        self.offset = 0

    def read_int(self) -> int:
        if self.offset + 4 > len(self.data):
            return 0
        value = int.from_bytes(self.data[self.offset:self.offset + 4], byteorder="big", signed=False)
        self.offset += 4
        return value

    def read_byte(self) -> int:
        if self.offset + 1 > len(self.data):
            return 0
        value = self.data[self.offset]
        self.offset += 1
        return value

    def read_string(self, length: int) -> str:
        if self.offset + length > len(self.data):
            return ""
        value = self.data[self.offset:self.offset + length]
        self.offset += length
        return value.decode("utf-8", errors="ignore")


class PacketPrinter:
    def __init__(self, debug: bool = False):
        self.debug = debug
        self.partial_segments = defaultdict(bytearray)
        self.loss_state = {}
        self.lock = threading.Lock()

    def handle_packet(self, packet) -> None:
        if TCP not in packet or Raw not in packet:
            return

        payload = bytes(packet[Raw].load)
        if len(payload) < 8:
            return

        tcp = packet[TCP]
        src, dst = self._get_endpoint(packet, tcp)
        if not src or not dst:
            return

        stream_key = (src, dst)
        buffer = self._reassemble_payload(stream_key, payload, bool(tcp.flags & 0x08))
        if not buffer:
            return

        self._print_payload(buffer, src, dst, int(tcp.dport), stream_key)

    def _get_endpoint(self, packet, tcp) -> Tuple[str, str]:
        if IP in packet:
            return f"{packet[IP].src}:{int(tcp.sport)}", f"{packet[IP].dst}:{int(tcp.dport)}"
        if IPv6 in packet:
            return f"{packet[IPv6].src}:{int(tcp.sport)}", f"{packet[IPv6].dst}:{int(tcp.dport)}"
        return "", ""

    def _reassemble_payload(self, stream_key: Tuple[str, str], payload: bytes, push_flag: bool) -> bytes:
        with self.lock:
            if not push_flag:
                self.partial_segments[stream_key].extend(payload)
                return b""

            if self.partial_segments[stream_key]:
                data = bytes(self.partial_segments[stream_key]) + payload
                self.partial_segments[stream_key].clear()
                return data

            return payload

    def _print_payload(self, buffer: bytes, src: str, dst: str, dst_port: int, stream_key: Tuple[str, str]) -> None:
        reader = BufferReader(buffer)
        buf_size = reader.read_int()

        if dst_port == 8001:
            reader.read_int()
            reader.read_int()
            reader.read_string(32)

        cmd_id = reader.read_int()

        if dst_port == 8001:
            reader.read_int()
            reader.read_int()
            reader.read_byte()
            key_index = reader.offset - 2
            if 0 <= key_index < len(reader.data):
                xor_key = reader.data[key_index]
                raw = bytes(value ^ xor_key for value in reader.data[reader.offset:])
            else:
                raw = reader.data[reader.offset:]
            print('123')
            if cmd_id == 6:
                with open("debug_type5.txt", "ab") as f:
                    f.write(reader.data[:reader.offset])
            self._emit(src, dst, cmd_id, 5, raw)
            return

        if len(buffer) <= 14:
            return

        data_type = buffer[12]
        if data_type == 3:
            if len(buffer) - buf_size != 4:
                with self.lock:
                    self.loss_state[stream_key] = LossState(cmd_id=cmd_id, needed_buf_size=buf_size, payload=buffer)
                return
            self._emit(src, dst, cmd_id, 3, self._parse_zlib_data(buffer[17:]))
            return

        if data_type == 5:
            if cmd_id == 6:
                print(f"Debug: type5 payload from {src} to {dst}: {buffer[12:]}")
                with open("debug_type51.txt", "ab") as f:
                    f.write(buffer[:12])
            self._emit(src, dst, cmd_id, 5, self._decode_type5(buffer[12:]))
            return

        if data_type == 2:
            self._emit(src, dst, cmd_id, 2, buffer[13:])
            return

        with self.lock:
            pending = self.loss_state.get(stream_key)

        if cmd_id > 99999 and pending is not None:
            combined = pending.payload + buffer
            if len(combined) - pending.needed_buf_size != 4:
                with self.lock:
                    self.loss_state[stream_key] = LossState(
                        cmd_id=pending.cmd_id,
                        needed_buf_size=pending.needed_buf_size,
                        payload=combined,
                    )
                return

            with self.lock:
                self.loss_state.pop(stream_key, None)
            self._emit(src, dst, pending.cmd_id, 3, self._parse_zlib_data(combined[17:]))

    def _emit(self, src: str, dst: str, cmd_id: int, data_type: int, payload: Union[bytes, str]) -> None:
        if isinstance(payload, bytes):
            message = payload.decode("utf-8", errors="replace")
        else:
            message = payload

        now = time.strftime("%Y-%m-%d %H:%M:%S")
        if cmd_id == 6 and data_type == 5:
            
            print(f"[{now}] {src} -> {dst} cmd={cmd_id} type={data_type} (type5 decoded)")
            print(message)

        #print(f"[{now}] {src} -> {dst} cmd={cmd_id} type={data_type}")
        #print(message)
        #print("-" * 80)

    @staticmethod
    def _decode_type5(data: bytes) -> bytes:
        if data and data[0] == 5:
            return bytes(value ^ 152 for value in data[1:])
        return b""

    @staticmethod
    def _parse_zlib_data(data: bytes) -> bytes:
        if len(data) >= 2 and data[0] == 120 and data[1] == 156:
            try:
                return zlib.decompress(data)
            except zlib.error:
                return data
        return data


@dataclass
class InterfaceEntry:
    index: int
    iface: object
    label: str
    aliases: Set[str]


def build_interface_entries() -> List[InterfaceEntry]:
    entries = []
    for index, iface in enumerate(get_working_ifaces(), start=1):
        name = getattr(iface, "name", "") or "<unknown>"
        description = getattr(iface, "description", "") or ""
        network_name = getattr(iface, "network_name", "") or ""
        label = f"{name} ({description})" if description else name
        aliases = {str(index), name.lower()}
        if network_name:
            aliases.add(network_name.lower())
        entries.append(InterfaceEntry(index=index, iface=iface, label=label, aliases=aliases))
    return entries


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="捕获 TCP 8001 流量并直接打印解析后的内容")
    parser.add_argument(
        "--iface",
        action="append",
        default=[],
        help="指定要监听的网卡，可传索引、名称或 network_name；不传则监听所有可用网卡",
    )
    parser.add_argument("--list-ifaces", action="store_true", help="只列出可用网卡，不开始抓包")
    parser.add_argument("--debug", action="store_true", help="打印额外调试信息")
    return parser.parse_args()


def select_interfaces(entries: List[InterfaceEntry], selected: List[str]) -> List[InterfaceEntry]:
    if not selected:
        return entries

    wanted = {item.lower() for item in selected}
    result = [entry for entry in entries if entry.aliases & wanted]
    if not result:
        raise SystemExit("未匹配到可用网卡，请先使用 --list-ifaces 查看列表")
    return result


def main() -> int:
    args = parse_args()
    entries = build_interface_entries()
    if not entries:
        print("未找到可用网卡，请确认已安装 Npcap。", file=sys.stderr)
        return 1

    if args.list_ifaces:
        print("可用网卡:")
        for entry in entries:
            print(f"{entry.index}: {entry.label}")
        return 0

    targets = select_interfaces(entries, args.iface)
    printer = PacketPrinter(debug=args.debug)
    sniffers = []
    stop_event = threading.Event()

    def stop_handler(*_args):
        stop_event.set()

    signal.signal(signal.SIGINT, stop_handler)
    if hasattr(signal, "SIGTERM"):
        signal.signal(signal.SIGTERM, stop_handler)

    for entry in targets:
        sniffer = AsyncSniffer(iface=entry.iface, filter="tcp and port 8001", prn=printer.handle_packet, store=False)
        sniffers.append((entry, sniffer))

    for entry, sniffer in sniffers:
        try:
            sniffer.start()
            print(f"开始抓包: {entry.index}: {entry.label}")
        except OSError as exc:
            print(f"启动网卡失败: {entry.label}: {exc}", file=sys.stderr)

    if not any(sniffer.running for _, sniffer in sniffers):
        print("没有可用的抓包任务启动成功。", file=sys.stderr)
        return 1

    print("按 Ctrl+C 停止。")
    try:
        while not stop_event.is_set():
            time.sleep(0.5)
    finally:
        for _, sniffer in sniffers:
            if sniffer.running:
                sniffer.stop()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())