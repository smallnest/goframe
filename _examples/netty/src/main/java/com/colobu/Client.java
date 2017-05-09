package com.colobu;

import io.netty.bootstrap.Bootstrap;
import io.netty.channel.*;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioSocketChannel;
import io.netty.handler.codec.LengthFieldBasedFrameDecoder;
import io.netty.handler.codec.LengthFieldPrepender;
import io.netty.handler.codec.string.StringDecoder;
import io.netty.handler.codec.string.StringEncoder;

public class Client {
    static final String HOST = System.getProperty("host", "127.0.0.1");
    static final int PORT = Integer.parseInt(System.getProperty("port", "9981"));

    public static void main(String[] args) throws InterruptedException {
        EventLoopGroup group = new NioEventLoopGroup();
        try {
            Bootstrap b = new Bootstrap();
            b.group(group)
                    .channel(NioSocketChannel.class)
                    .option(ChannelOption.TCP_NODELAY, true)
                    .handler(new ChannelInitializer<SocketChannel>() {
                        @Override
                        public void initChannel(SocketChannel ch) throws Exception {
                            ChannelPipeline p = ch.pipeline();

                            p.addLast("frameDecoder",
                                    new LengthFieldBasedFrameDecoder(1048576, 0, 4, 0, 4));
                            p.addLast("frameEncoder", new LengthFieldPrepender(4));
                            p.addLast(new StringDecoder());
                            p.addLast(new StringEncoder());

                            //p.addLast(new LoggingHandler(LogLevel.INFO));
                            p.addLast(new ClientHandler());
                        }
                    });

            // Start the client.
            ChannelFuture f = b.connect(HOST, PORT).sync();
            Channel ch = f.channel();
            ch.writeAndFlush("hello");
            ch.writeAndFlush("world");
            
            // Wait until the connection is closed.
            f.channel().closeFuture().sync();
        } finally {
            // Shut down the event loop to terminate all threads.
            group.shutdownGracefully();
        }

    }
}
