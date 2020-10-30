#version 330

in vec2 in_position;
in vec2 in_mid;
in vec2 in_tex_coord;
in float hue;

uniform mat4 proj;
uniform float scale;
uniform float points;
uniform float endScale;
uniform float hueshift;
uniform float saturation;
uniform float instances;

out vec2 tex_coord;
out vec4 color_pass;
out float index;

vec3 hsv2rgb(vec3 c) {
    vec4 K = vec4(1.0, 2.0 / 3.0, 1.0 / 3.0, 3.0);
    vec3 p = abs(fract(c.xxx + K.xyz) * 6.0 - K.www);
    return c.z * mix(K.xxx, clamp(p - K.xxx, 0.0, 1.0), c.y);
}

void main() {
    gl_Position = proj * vec4(in_position * scale * mix(endScale, 1, smoothstep(instances - points, instances, gl_InstanceID)) + in_mid, 0.0, 1.0);
    tex_coord = in_tex_coord;
    index = gl_InstanceID;
    color_pass = vec4(hsv2rgb(vec3(fract(hue + hueshift), saturation, 1.0)), 1.0);
}